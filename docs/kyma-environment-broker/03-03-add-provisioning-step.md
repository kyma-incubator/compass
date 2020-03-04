# Configure Runtime provisioning

You can configure Runtime provisioning process by providing additional input objects in the form of overrides. For example, you may want to provide tokens/credentials/URLs to integrate Kyma Runtime with external systems. 

The operation of provisioning a Runtime consists of several steps. Each step is represented by a file that is responsible for a separate part of preparing Runtime parameters. The last step is called [`create_runtime`](https://github.com/kyma-incubator/compass/blob/master/components/kyma-environment-broker/internal/process/provisioning/create_runtime.go). It transforms all steps into a request to provision a Runtime. The request is sent to the Runtime Provisioner component.

In case of a processing failure, every step can be re-launched multiple times. That is why, for each step you should determine a behavior in case of a processing failure. It can either:
- Return an error, which interrupts the entire provisioning process, or 
- Repeat the entire operation after the specified period. 

## Add provisioning step

To add a new provisioning step, follow these steps:

1. Create a new file in [this](https://github.com/kyma-incubator/compass/blob/master/components/kyma-environment-broker/internal/process/provisioning) directory. 

2. Implement this interface in your provisioning step:

    ```go
    type Step interface {
        Name() string
        Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error)
    }
    ```

    - `Name()` method returns the name of the step that is used in logs.
    - `Run()` method implements the functionality of the step. The method receives operations as an argument to which it can add appropriate overrids or save other used variables.
    

    ```go
    operation.InputCreator.SetOverrides(COMPONENT_NAME, []*gqlschema.ConfigEntryInput{
        {
            Key:   "path.to.key",
            Value: SOME_VALUE,
        },
        {
            Key:    "path.to.secret",
            Value:  SOME_VALUE,
            Secret: ptr.Bool(true),
        },
    })
    ```

    If your functionality contains long-term processes, you can store data in the storage.
    To do this, add this field to the provisioning operation in which you want to save data:

    ```go
    type ProvisioningOperation struct {
        Operation `json:"-"`
    
        // These fields are serialized to JSON and stored in the storage
        LmsTenantID            string `json:"lms_tenant_id"`
        ProvisioningParameters string `json:"provisioning_parameters"`
    
        NewFieldFromCustomStep string `json:"new_field_from_custom_step"`    
    
        // These fields are not stored in the storage
        InputCreator ProvisionInputCreator `json:"-"`
    }
    ```

    By saving data in the storage, you can check if you already have the necessary data and avoid time-consuming processes.       You should always return the modified operation from the method. 

    See the example of the step implementation:

    ```go
    package provisioning
    
    import (
        "encoding/json"
        "net/http"
        "time"
    
        "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
        "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
        "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
    
        "github.com/sirupsen/logrus"
    )
    
    type HelloWorldStep struct {
        operationStorage storage.Operations
        client           *http.Client
    }
    
    type ExternalBodyResponse struct {
        data  string
        token string
    }
    
    func NewHelloWorldStep(operationStorage storage.Operations, client *http.Client) *HelloWorldStep {
        return &HelloWorldStep{
            operationStorage: operationStorage,
            client:           client,
        }
    }
    
    func (s *HelloWorldStep) Name() string {
        return "Hello_World"
    }
    
    // Your step can be repeated in case any other step fails, even if your step has already done its job
    func (s *HelloWorldStep) Run(operation internal.ProvisioningOperation, log *logrus.Entry) (internal.ProvisioningOperation, time.Duration, error) {
        log.Info("Start step")
   
        // check if step should run or his job is done in the previous iteration
        // All non-save operation data are empty (e.g. InputCreator overrides)
    
        // Add your logic here
    
        // Add a call to an external service
        response, err := s.client.Get("http://example.com")
        if err != nil {
            // Error during a call to an external service may be temporary so you should return time.Duration 
            // All steps will be repeated in X seconds/minutes
            return operation, 1 * time.Second, nil
        }
        defer response.Body.Close()
    
        body := ExternalBodyResponse{}
        err = json.NewDecoder(response.Body).Decode(&body)
        if err != nil {
            log.Errorf("error: %s", err)
            // Handle a process failure by returning an error or time.Duration
        }
    
        // If a call or any other action is time-consuming, you can save the result in the operation
        // If you need an extra field in the ProvisioningOperation structure, add it first
        // in the step below; beforehand, you can check if a given value already exists in the operation
        operation.HelloWorlds = body.data
        updatedOperation, err := s.operationStorage.UpdateProvisioningOperation(operation)
        if err != nil {
            log.Errorf("error: %s", err)
            // Handle a process failure by returning an error or time.Duration
        }
    
        // If your step finishes with data which should be added to override used during the Runtime provisioning,
        // add an extra value to operation.InputCreator, then return the updated version of the Application
        updatedOperation.InputCreator.SetOverrides("component-name", []*gqlschema.ConfigEntryInput{
            {
                Key:   "some.key",
                Value: body.token,
            },
        })
    
        // Return the updated version of the Application
        return *updatedOperation, 0, nil
    }
    ```

3. Add the step to the [`/cmd/broker/main.go`](https://github.com/kyma-incubator/compass/blob/master/components/kyma-environment-broker/cmd/broker/main.go) file:

    ```go
    helloWorldStep := provisioning.NewHelloWorldStep(db.Operations(), &http.Client{})
    (...)
    stepManager.AddStep(1, helloWorldStep)
    ```

    The weight of the step should be greater than or equal to 1. If you want the step to be performed before a call to the Runtime Provisioner, its weight must be lower than `runtimeStep`.
