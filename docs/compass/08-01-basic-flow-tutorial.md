# Manage your Applications using Kyma Console and Compass UIs

This tutorial presents the basic flow in which you manually register an external Application's API into Compass and expose it into the Kyma Runtime. In the Kyma Runtime, you then create a Function that calls the Application's API. While going through this tutorial, you will navigate between two UI views:

- Compass UI, where you create connections between Applications, Runtimes, and scenarios
- Kyma Console UI, where you manage resources used in your Application, such as services, Functions, and bindings

## Prerequisites

For simplicity reasons, use the available Order Service as the sample external Application for this tutorial. Prepare the following:

- [`order-service`](./assets/order-service.yaml) file that contains the service definition, deployment, and its API
- [API specification](./assets/order-service-api-spec.yaml) of `order-service`
- [Function](./assets/function.yaml) that calls `order-service` for orders
- [Compass cluster with all Kyma components installed](./04-01-installation.md#single-cluster-with-compass-and-runtime-agent). This tutorial won't work on local k3d installation.

>**NOTE:** Read [this](./04-01-installation.md) document to learn how to install Kyma with the Compass module.

## Steps

### Deploy the external Application

1. Log in to the Kyma Console and create a new Namespace by selecting the **Add new namespace** button.

2. In the **Overview** tab in your Namespace, select the **Deploy new resource** button and use the [`order-service`](./assets/order-service.yaml) file to connect the Application.

3. In the **API Rules** tab, copy the URL to your Application. You will use it in the Compass UI in the next steps.

### Register your Application in the Compass UI

1. Open a separate tab in your browser and go to `https://compass.{CLUSTER_DOMAIN}`. It will navigate you to the Compass UI. From the drop-down list on the top navigation panel, select the tenant you want to work on. For the purpose of this tutorial, select the `default` tenant. In the **Runtimes** tab, there is already the default `kymaruntime` that you can work on to complete this tutorial. Make sure that your Runtime is assigned to the `DEFAULT` scenario.

2. In the left navigation panel, navigate to the **Application** tab and click **Create application...** to register your Application in Compass. Choose **From scratch** from the drop-down list. For the purpose of this tutorial, name your Application `test-app`. By default, your Application is assigned to the `DEFAULT` scenario. Click the **Create** button.

3. Select `test-app` in the **Applications** view and add the API spec of the Order Service:

    a. Click the **Create Bundle** button in the **Bundles** section and fill in all the required fields. For the purpose of this tutorial, name your Bundle `test-bundle`.

    b. In the **Credentials** tab, choose `Empty` from the drop-down list to select the credentials type. For the purpose of this tutorial, there is no need to secure the connection. Click **Create**.

    c. Navigate to the `test-bundle` Bundle and click the **+** button in the **API Definitions** section. Fill in all the required fields. In the **Target URL** field, paste the URL to your Application.

    d. Click the **Add specification** button and upload the `order-service` [API spec file](./assets/order-service-api-spec.yaml). Click **Create**.


### Use your Application in the Kyma Console UI

1. Go back to the Kyma Console UI. You can see that the `test-app` Application is registered in the **Applications/Systems** view. Select `test-app` and bind it to your Namespace by selecting the **Create Binding** button.

2. From the drop-down list in the top right corner, select your Namespace and go to the **Catalog** view. You will see your services available under the **Services** tab. Select the `test-app` service. Provision the service instance by choosing your Bundle as **Selected plan** from the dropdown list and clicking the **Add** button in the top right corner of the page.

3. Create a Function. In the **Overview** tab, click the **Deploy new resource** button and upload the file with the [Function](./assets/function.yaml).

4. Bind your Function:

    a. In the left navigation panel, go to the **Functions** tab and click the `call-order-service` Function.

    b. Scroll down to the end of the **Configuration** view and bind your Function to your instance by clicking the **Create Service Binding** button in the **Service Binding** section. Choose the ServiceInstance you want to bind your function to and click **Create Service Binding**.

5. Expose your Function:

    a. In the left navigation panel, go to the **API Rules** tab and click the **Create API Rule** button.

    b. Type `test-function` in both the **Name** and **Hostname** fields. From the **Service** dropdown, select `call-order-service`. Keep **Access strategies** as they are. Click the **Create** button in the top right corner of the page.

    >**CAUTION:** If you use the `xip.io` domain with a self-signed certificate, you must edit the Gateway deployment to be able to use your Function. To do so, go to the **Deployments** tab in the left navigation panel, unfold the vertical option menu in the `{YOUR_NAMESPACE}-gateway` row, and click **Edit**. Scroll down to line 52 and set the **--skipVerify** flag to `true`. Click **Update**.

6. Test your Function. Open a separate tab in your browser and go to `https://test-function.{CLUSTER_DOMAIN}`. It will take you to your Function view that returns the following output:

    ```
    {
      "status": 200,
      "data": []
    }
    ```

    You can play with your Function by performing the following actions in the terminal window:

    - `{"action":"add"}` - adds the new order
    - `{"action":"list"}` - lists all orders (this is the default command executed after you click the **Send** button)
    - `{"action":"delete"}` - deletes all the orders

    To perform an action, run the following command:

    ```bash
    curl --header "Content-Type: application/json" --request POST --data '{SELECTED_ACTION}' https://test-function.{CLUSTER_DOMAIN}
    ```

### Cleanup

Clean up your cluster after going through this tutorial. To do so, delete your resources in the following order:

1. Go to the **Functions** tab and delete your Function.

2. Go to the **Services** tab and delete your services.

3. Go to the **Deployments** tab and delete your deployments.

4. Go to the **API Rules** tab and delete your APIRules.

5. Go to the **Instances** tab and delete your instances.

6. Go to the **Overview** section and unbind `test-app` from your Namespace.

7. Go back to the Namespaces view and delete your Namespace.

8. In the Compass UI, remove `test-app` from the **Applications** view. If you go back to the **Applications/Systems** view in the Kyma Console UI, you can see that the `test-app` Application is removed. It can take a moment, so refresh the Console UI if the Application is still there.
