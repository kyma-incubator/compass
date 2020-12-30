package com.sap.cloud.cmp.ord.service.storage.model.converter;

import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

import static org.junit.Assert.assertEquals;

@RunWith(SpringRunner.class)
@SpringBootTest(classes = UrlConverter.class)
public class UrlConverterTest {

    @Value("${server.self_url}")
    private String serverUrl;

    @Value("${odata.jpa.request_mapping_path}")
    private String requestMappingPath;


    @Test
    public void testConvertToEntityAttribute_ReturnsAbsoluteUrl_WhenInvoked() {
        UrlConverter urlConverter = new UrlConverter();

        String path = "/v1/api";

        String actualUrl = urlConverter.convertToEntityAttribute(path);
        String expectedUrl = serverUrl + "/" + requestMappingPath + path;

        assertEquals(expectedUrl, actualUrl);
    }
}
