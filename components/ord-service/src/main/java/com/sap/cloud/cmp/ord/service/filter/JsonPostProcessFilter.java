package com.sap.cloud.cmp.ord.service.filter;

import com.sap.cloud.cmp.ord.service.filter.wrappers.JsonResponseWrapper;
import org.springframework.stereotype.Component;

import javax.servlet.*;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;

@Component
public class JsonPostProcessFilter implements Filter {

    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain filterChain) throws IOException, ServletException {
        JsonResponseWrapper capturingResponseWrapper = new JsonResponseWrapper((HttpServletResponse) response);

        filterChain.doFilter(request, capturingResponseWrapper);

        if (response.getContentType() != null && response.getContentType().contains("application/json")) {
            String content = capturingResponseWrapper.getCaptureAsString();

            content = content.replaceAll("\\\\\"","\"").replaceAll("\"\\{","{").replaceAll("}\"", "}");

            response.setContentLength(content.length());
            response.getWriter().write(content);
        }
    }
}
