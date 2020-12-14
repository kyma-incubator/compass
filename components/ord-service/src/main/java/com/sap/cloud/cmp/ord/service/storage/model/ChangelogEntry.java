package com.sap.cloud.cmp.ord.service.storage.model;

import javax.persistence.Column;
import javax.persistence.Embeddable;

@Embeddable
public class ChangelogEntry {
    @Column(name = "version", length = Integer.MAX_VALUE)
    private String version;

    @Column(name = "release_status", length = Integer.MAX_VALUE)
    private String releaseStatus;

    @Column(name = "date", length = Integer.MAX_VALUE)
    private String date;

    @Column(name = "description", length = Integer.MAX_VALUE)
    private String description;

    @Column(name = "url", length = Integer.MAX_VALUE)
    private String url;
}
