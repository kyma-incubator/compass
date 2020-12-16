package com.sap.cloud.cmp.ord.service.storage.model;

import javax.persistence.Column;
import javax.persistence.Embeddable;

@Embeddable
public class ArrayElement {

    @Column(name = "value", length = Integer.MAX_VALUE)
    private String value;
}
