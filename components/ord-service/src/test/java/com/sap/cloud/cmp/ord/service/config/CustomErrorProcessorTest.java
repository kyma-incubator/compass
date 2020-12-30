package com.sap.cloud.cmp.ord.service.config;

import org.apache.olingo.server.api.ODataServerError;
import org.junit.Test;

import static org.junit.Assert.assertEquals;


public class CustomErrorProcessorTest {

    @Test
    public void testProcessError_ReturnsAppropriateMessage_WhenInvoked() {
        String errorMessage = "Server error occurred.";
        ODataServerError oDataServerError = new ODataServerError();
        oDataServerError.setMessage(errorMessage);

        CustomErrorProcessor customErrorProcessor = new CustomErrorProcessor();
        customErrorProcessor.processError(null, oDataServerError);

        assertEquals(errorMessage + CustomErrorProcessor.ADDITIONAL_ERR_MESSAGE, oDataServerError.getMessage());
    }
}
