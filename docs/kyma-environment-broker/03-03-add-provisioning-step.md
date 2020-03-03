# Provisioning runtime details

You can configure Runtime provisioning process by providing additional input parameters in the form of overrides. For example, you may want to provide tokens/credentials/URLs... <example here>. 

The operation of provisioning a Runtime consists of steps. Each step is represented by a file that is responsible for a separate part of preparing Runtime parameters. The last step is [`create_runtime`](https://github.com/kyma-incubator/compass/blob/master/components/kyma-environment-broker/internal/process/provisioning/create_runtime.go) which transforms all steps into a request to provision a Runtime that is sent to the Runtime Provisioner component.

In case of an error, every step can be re-launched multiple times even if it has already been processed before.

For each step, you should determine a behavior in case of a processing failure. It can either:
- Return an error, which interrupts the entire provisioning process, or 
- Repeat the entire operation after the specified period. 

## Add a custom step

1. Each step has to implement the interface

    ```go
    type Step interface {
        Name() string
        Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error)
    }
    ```

2. `Name()` method should return step name which will be use in logs.

3. `Run()` method has to implement all step functionality. The method receives 
operations as an argument to which it can add appropriate overrids or save other used variables.

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

    If your functionality contains long-term processes, some data can be stored in the storage in a specific operation.
    To do this you need to add a field to the operation to which you will be saving data.

    ```go
    type ProvisioningOperation struct {
        Operation `json:"-"`
    
        // following fields are serialized to JSON and stored in the storage
        LmsTenantID            string `json:"lms_tenant_id"`
        ProvisioningParameters string `json:"provisioning_parameters"`
    
        NewFieldFromCustomStep string `json:"new_field_from_custom_step"`    
    
        // following fields are not stored in the storage
        InputCreator ProvisionInputCreator `json:"-"`
    }
    ```

    Thanks to this approach, when performing your operation again you can check if you already have the 
    necessary data and avoid time-consuming processes.

    The modified operation should be returned from the method. 
    If the operation (not the step) should be interrupted, the method should return an error.

    Below is an example of a step implementation:

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
    
    // remember your step could be repeated if any other step will fail even if your step done his job
    func (s *HelloWorldStep) Run(operation internal.ProvisioningOperation, log *logrus.Entry) (internal.ProvisioningOperation, time.Duration, error) {
        log.Info("Start step")
   
        // check if step should run or his job is done in the previous iteration
        // remember all non save in storage operation data are empty (e.g. InputCreator overrides)
    
        // prepare data
    
        // call to external services
        response, err := s.client.Get("http://example.com")
        if err != nil {
            // err during call to external service may be temporary so time.Duration should be returned
            // what means all operation process (all steps) will be repeated in X second/minute...
            return operation, 1 * time.Second, nil
        }
        defer response.Body.Close()
    
        body := ExternalBodyResponse{}
        err = json.NewDecoder(response.Body).Decode(&body)
        if err != nil {
            log.Errorf("error: %s", err)
            // handle error by returning error or time.Duration
        }
    
        // if call or any other action is time-consuming then result can be saved in operation
        // if you need extra field in ProvisioningOperation struct, add it first
        // in step below first you can check if value already exist in operation
        operation.HelloWorlds = body.data
        updatedOperation, err := s.operationStorage.UpdateProvisioningOperation(operation)
        if err != nil {
            log.Errorf("error: %s", err)
            // handle error by returning error or time.Duration
        }
    
        // if your step finish with data wich should be add to override which will be used during runtime provisioning
        // add extra value to operation.InputCreator, then return updated version of application
        updatedOperation.InputCreator.SetOverrides("component-name", []*gqlschema.ConfigEntryInput{
            {
                Key:   "some.key",
                Value: body.token,
            },
        })
    
        // return updated version of application
        return *updatedOperation, 0, nil
    }
    ```

4. Add the finished step to the `/cmd/broker/main.go` file.

    ```go
    helloWorldStep := provisioning.NewHelloWorldStep(db.Operations(), &http.Client{})
    (...)
    stepManager.AddStep(1, helloWorldStep)
    ```

    The weight of the step should be greater than or equal to 1.
    If the step should be made before call to the provisioner (which runs the Runtime process creation) 
    its weight should be lower than `runtimeStep`
