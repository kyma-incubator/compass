package com.sap.cloud.cmp.ord.service.config;

import com.sap.olingo.jpa.processor.core.api.JPAErrorProcessor;
import org.apache.olingo.server.api.*;

public class CustomErrorProcessor implements JPAErrorProcessor {

    public final static String ADDITIONAL_ERR_MESSAGE =  " Use odata-debug query parameter with value one of the following formats: json,html,download for more information.";

    @Override
    public void processError(ODataRequest oDataRequest, ODataServerError oDataServerError) {
        oDataServerError.setMessage(oDataServerError.getMessage() + ADDITIONAL_ERR_MESSAGE);
    }
}
