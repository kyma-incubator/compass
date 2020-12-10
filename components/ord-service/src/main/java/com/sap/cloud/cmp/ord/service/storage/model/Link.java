package com.sap.cloud.cmp.ord.service.storage.model;

public class Link {
    private String title;
    private String description;
    private String url;
    private String extensions;

    public Link() {}

    public Link(String title, String description, String url, String extensions) {
        this.title = title;
        this.description = description;
        this.url = url;
        this.extensions = extensions;
    }

    public String getTitle() {
        return title;
    }

    public void setTitle(String title) {
        this.title = title;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public String getUrl() {
        return url;
    }

    public void setUrl(String url) {
        this.url = url;
    }

    public String getExtensions() {
        return extensions;
    }

    public void setExtensions(String extensions) {
        this.extensions = extensions;
    }
}
