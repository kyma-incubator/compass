package com.sap.cloud.cmp.ord.service.filter;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.sap.cloud.cmp.ord.service.filter.wrappers.JsonResponseWrapper;
import org.springframework.stereotype.Component;

import javax.servlet.*;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.util.*;

@Component
public class JsonPostProcessFilter implements Filter {

    private static final String[] subresources = new String[]{"value" /*This is the key returned when you list objects*/,"products", "packages","apis","events"};
    private static final String[] arrays = new String[]{"tags", "countries","lineOfBusiness","industry"};

    private static final ObjectMapper mapper = new ObjectMapper();

    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain filterChain) throws IOException, ServletException {
        JsonResponseWrapper capturingResponseWrapper = new JsonResponseWrapper((HttpServletResponse) response);

        filterChain.doFilter(request, capturingResponseWrapper);

        String content = capturingResponseWrapper.getCaptureAsString();

        if (response.getContentType() != null && response.getContentType().contains("application/json")) {
            // Make JSON returned as String to look like real JSON
            // content = content.replaceAll("\\\\\"","\"").replaceAll("\"\\{","{").replaceAll("}\"", "}");

            // Aggreagate Array Elements
            if (Boolean.TRUE.toString().equals(request.getParameter("compact"))) {
                JsonNode jsonTree = mapper.readTree(content);
                aggregateArrayElements(jsonTree);
                content = mapper.writeValueAsString(jsonTree);
            }
        }

        response.setContentLength(content.length());
        response.getWriter().write(content);
    }

    private void aggregateArrayElements(JsonNode jsonTree) {
        if (jsonTree != null) {
            if (jsonTree.isArray()) {
                Iterator<JsonNode> it = jsonTree.elements();
                while (it.hasNext()) {
                    JsonNode el = it.next();
                    processResource(el);
                    for (String subresource : subresources) {
                        aggregateArrayElements(el.get(subresource));
                    }
                }
            } else {
                processResource(jsonTree);
                for (String subresource : subresources) {
                    aggregateArrayElements(jsonTree.get(subresource));
                }
            }
        }
    }

    private void processResource(JsonNode resource) {
        for (String arrayName : arrays) {
            JsonNode array = resource.get(arrayName);
            if (array != null && array.isArray()) {
                JsonNode convertedArray = convertArray(array);
                ((ObjectNode) resource).replace(arrayName, convertedArray);
            }
        }

        JsonNode labels = resource.get("labels");
        if (labels != null && labels.isArray()) {
            JsonNode convertedLabels = convertLabels(labels);
            ((ObjectNode) resource).replace("labels", convertedLabels);
        }
    }

    private JsonNode convertArray(JsonNode array) {
        if (array == null || !array.isArray()) {
            return array;
        }
        List<String> result = new ArrayList<>();
        Iterator<JsonNode> it = array.elements();
        while(it.hasNext()) {
            JsonNode el = it.next();
            JsonNode value = el.get("value");
            if (value != null) {
                result.add(value.textValue());
            }
        }
        return mapper.valueToTree(result);
    }

    private JsonNode convertLabels(JsonNode labels) {
        if (labels == null || !labels.isArray()) {
            return labels;
        }
        Map<String,List<String>> result = new HashMap<>();
        Iterator<JsonNode> it = labels.elements();
        while(it.hasNext()) {
            JsonNode el = it.next();
            JsonNode key = el.get("key");
            JsonNode value = el.get("value");
            if (key != null && value != null) {
                List<String> label = result.get(key.textValue());
                if (label == null) {
                    label = new ArrayList<>();
                }
                label.add(value.textValue());
                result.put(key.textValue(), label);
            }
        }
        return mapper.valueToTree(result);
    }
}
