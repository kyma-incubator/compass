package com.sap.cloud.cmp.ord.service.storage.model;

import javax.persistence.Column;
import javax.persistence.Embeddable;

@Embeddable
public class Tag {
    @Column(name = "key", length = Integer.MAX_VALUE)
    private String key;

    @Column(name = "value", length = Integer.MAX_VALUE)
    private String value;
}
