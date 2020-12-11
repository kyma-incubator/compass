package com.sap.cloud.cmp.ord.service.storage.model;

import javax.persistence.Column;
import javax.persistence.Embeddable;

@Embeddable
public class Provider {
    @Column(name = "name", length = Integer.MAX_VALUE)
    private String name;

    @Column(name = "department", length = Integer.MAX_VALUE)
    private String department;

    @Column(name = "extensions", length = Integer.MAX_VALUE)
    private String extensions;
}
