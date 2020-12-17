package com.sap.cloud.cmp.ord.service.controller;


import com.sap.cloud.cmp.ord.service.repository.ApiSpecRepository;
import com.sap.cloud.cmp.ord.service.repository.EventSpecRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.ResponseBody;

import java.util.UUID;

@Controller
public class SpecificationsController {

    @Autowired
    private ApiSpecRepository apiSpecRepository;

    @Autowired
    private EventSpecRepository eventSpecRepository;


    @RequestMapping(value = "/${odata.jpa.request_mapping_path}/api/{id}/specification", method = { RequestMethod.GET }, produces = MediaType.TEXT_PLAIN_VALUE)
    @ResponseBody
    public String getApiSpec(@PathVariable final String id) {
        return apiSpecRepository.getOne(UUID.fromString(id)).getSpecData();
    }

    @RequestMapping(value = "/${odata.jpa.request_mapping_path}/event/{id}/specification", method = { RequestMethod.GET }, produces = MediaType.TEXT_PLAIN_VALUE)
    @ResponseBody
    public String getEventSpec(@PathVariable final String id) {
        return eventSpecRepository.getOne(UUID.fromString(id)).getSpecData();
    }

}
