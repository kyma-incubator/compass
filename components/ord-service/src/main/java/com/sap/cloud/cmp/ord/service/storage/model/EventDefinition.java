package com.sap.cloud.cmp.ord.service.storage.model;

import javax.persistence.Column;
import javax.persistence.Embeddable;

@Embeddable
public class EventDefinition {
    @Column(name = "type", length = Integer.MAX_VALUE)
    private String type;

    @Column(name = "custom_type", length = Integer.MAX_VALUE)
    private String customType;

    @Column(name = "media_type", length = Integer.MAX_VALUE)
    private String mediaType;

    @Column(name = "url", length = Integer.MAX_VALUE)
    private String url;
}
