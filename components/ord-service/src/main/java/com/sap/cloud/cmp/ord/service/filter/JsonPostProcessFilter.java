package com.sap.cloud.cmp.ord.service.filter;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.sap.cloud.cmp.ord.service.filter.aggregator.JsonArrayElementsAggregator;
import com.sap.cloud.cmp.ord.service.filter.wrappers.JsonResponseWrapper;
import org.springframework.stereotype.Component;

import javax.servlet.*;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;

@Component
public class JsonPostProcessFilter implements Filter {

    private final static String COMPACT_QUERY_PARAM = "compact";

    private static final ObjectMapper mapper = new ObjectMapper();

    private static final JsonArrayElementsAggregator aggregator = new JsonArrayElementsAggregator(mapper);

    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain filterChain) throws IOException, ServletException {
        JsonResponseWrapper capturingResponseWrapper = new JsonResponseWrapper((HttpServletResponse) response);

        filterChain.doFilter(request, capturingResponseWrapper);

        String content = capturingResponseWrapper.getCaptureAsString();

        if (response.getContentType() != null && response.getContentType().contains("application/json")) {
            // Make JSON returned as String to look like real JSON
            // content = content.replaceAll("\\\\\"","\"").replaceAll("\"\\{","{").replaceAll("}\"", "}");

            // Aggreagate Array Elements
            if (Boolean.TRUE.toString().equals(request.getParameter(COMPACT_QUERY_PARAM))) {
                JsonNode jsonTree = mapper.readTree(content);
                aggregator.aggregate(jsonTree);
                content = mapper.writeValueAsString(jsonTree);
            }
        }

        response.setContentLength(content.length());
        response.getWriter().write(content);
    }

}
