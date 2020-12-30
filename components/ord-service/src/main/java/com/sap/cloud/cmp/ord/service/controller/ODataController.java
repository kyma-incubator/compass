package com.sap.cloud.cmp.ord.service.controller;

import com.sap.olingo.jpa.processor.core.api.JPAODataCRUDContextAccess;
import com.sap.olingo.jpa.processor.core.api.JPAODataGetHandler;
import org.apache.olingo.commons.api.ex.ODataException;
import org.apache.olingo.server.api.debug.DefaultDebugSupport;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

@Controller
@RequestMapping("/${odata.jpa.request_mapping_path}/**")
public class ODataController {

    @Autowired
    private JPAODataCRUDContextAccess serviceContext;

    @RequestMapping(value = "**", method = { RequestMethod.GET })
    public void handleODataRequest(HttpServletRequest request, HttpServletResponse response) throws ODataException {
        final JPAODataGetHandler handler = new JPAODataGetHandler(serviceContext);
        //handler.getJPAODataRequestContext().setDebugSupport(new DefaultDebugSupport()); // Use query parameter odata-debug=json to activate.
        // Activate debug support after security is in place

        handler.process(request, response);
    }
}
