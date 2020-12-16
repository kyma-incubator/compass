package com.sap.cloud.cmp.ord.service.storage.model;

import javax.persistence.Column;
import javax.persistence.Embeddable;

@Embeddable
public class APIResourceLink {
    @Column(name = "type", length = Integer.MAX_VALUE)
    private String type;

    @Column(name = "custom_type", length = Integer.MAX_VALUE)
    private String customType;

    @Column(name = "url", length = Integer.MAX_VALUE)
    private String url;
}
