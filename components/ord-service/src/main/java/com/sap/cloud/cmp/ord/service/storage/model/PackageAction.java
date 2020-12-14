package com.sap.cloud.cmp.ord.service.storage.model;

import javax.persistence.Column;
import javax.persistence.Embeddable;

@Embeddable
public class PackageAction {
    @Column(name = "target", length = Integer.MAX_VALUE)
    private String target;

    @Column(name = "type", length = Integer.MAX_VALUE)
    private String type;

    @Column(name = "custom_type", length = Integer.MAX_VALUE)
    private String customType;

    @Column(name = "description", length = Integer.MAX_VALUE)
    private String description;

    @Column(name = "extensions", length = Integer.MAX_VALUE)
    private String extensions;
}
