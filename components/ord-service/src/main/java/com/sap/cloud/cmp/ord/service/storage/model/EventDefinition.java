package com.sap.cloud.cmp.ord.service.storage.model;

import com.sap.cloud.cmp.ord.service.storage.model.converter.UrlConverter;

import javax.persistence.Column;
import javax.persistence.Convert;
import javax.persistence.Embeddable;

@Embeddable
public class EventDefinition {
    @Column(name = "type", length = Integer.MAX_VALUE)
    private String type;

    @Column(name = "custom_type", length = Integer.MAX_VALUE)
    private String customType;

    @Column(name = "media_type", length = Integer.MAX_VALUE)
    private String mediaType;

    @Convert(converter = UrlConverter.class)
    @Column(name = "url", length = Integer.MAX_VALUE)
    private String url;
}
