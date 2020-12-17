package com.sap.cloud.cmp.ord.service.storage.model.converter;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.core.env.Environment;
import org.springframework.stereotype.Component;

import javax.persistence.AttributeConverter;
import javax.persistence.Converter;

@Converter
@Component
public class UrlConverter implements AttributeConverter<String, String> {

    private static Environment env;

    @Override
    public String convertToDatabaseColumn(String s) {
        return null; // ORD Service is read only
    }

    @Override
    public String convertToEntityAttribute(String s) {
        return env.getProperty("server.self_url") + "/" + env.getProperty("odata.jpa.request_mapping_path") + s;
    }

    @Autowired
    public void setEnv(Environment env) {
        UrlConverter.env = env;
    }
}
