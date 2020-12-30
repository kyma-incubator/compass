package com.sap.cloud.cmp.ord.service.filter.aggregator;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.Test;

import static org.junit.Assert.assertEquals;

public class JsonArrayElementsAggregatorTest {

    private static final ObjectMapper mapper = new ObjectMapper();

    @Test
    public void testAggregate_ReturnsUnmodifiedJson_WhenArrayElementNotFound() throws Exception {
        JsonArrayElementsAggregator aggregator = new JsonArrayElementsAggregator(mapper);

        String content = "{\"@odata.context\":\"$metadata#apis\",\"value\":[{\"ordId\":\"test-id\"}]}";
        String expectedContent  = "{\"@odata.context\":\"$metadata#apis\",\"value\":[{\"ordId\":\"test-id\"}]}";

        JsonNode jsonTree = mapper.readTree(content);
        aggregator.aggregate(jsonTree);
        String actualContent = mapper.writeValueAsString(jsonTree);

        assertEquals(expectedContent, actualContent);
    }

    @Test
    public void testAggregate_ReturnsUnmodifiedJson_WhenArrayElementFoundButNotKnown() throws Exception {
        JsonArrayElementsAggregator aggregator = new JsonArrayElementsAggregator(mapper);

        String content = "{\"@odata.context\":\"$metadata#apis\",\"value\":[{\"unknownArray\":[{\"value\":\"automotive\"},{\"value\":\"finance\"}]}]}";
        String expectedContent  = "{\"@odata.context\":\"$metadata#apis\",\"value\":[{\"unknownArray\":[{\"value\":\"automotive\"},{\"value\":\"finance\"}]}]}";

        JsonNode jsonTree = mapper.readTree(content);
        aggregator.aggregate(jsonTree);
        String actualContent = mapper.writeValueAsString(jsonTree);

        assertEquals(expectedContent, actualContent);
    }

    @Test
    public void testAggregate_ReturnsModifiedJson_WhenArrayElementFound() throws Exception {
        JsonArrayElementsAggregator aggregator = new JsonArrayElementsAggregator(mapper);

        String content = "{\"@odata.context\":\"$metadata#apis\",\"value\":[{\"tags\":[{\"value\":\"automotive\"},{\"value\":\"finance\"}]}]}";
        String expectedContent = "{\"@odata.context\":\"$metadata#apis\",\"value\":[{\"tags\":[\"automotive\",\"finance\"]}]}";

        JsonNode jsonTree = mapper.readTree(content);
        aggregator.aggregate(jsonTree);
        String actualContent = mapper.writeValueAsString(jsonTree);

        assertEquals(expectedContent, actualContent);
    }

    @Test
    public void testAggregate_ReturnsUnmodifiedJson_WhenLabelsFound() throws Exception {
        JsonArrayElementsAggregator aggregator = new JsonArrayElementsAggregator(mapper);

        String content = "{\"@odata.context\":\"$metadata#apis\",\"value\":[{\"labels\":[{\"key\":\"country\",\"value\":\"DE\"},{\"key\":\"country\",\"value\":\"US\"}]}]}";
        String expectedContent = "{\"@odata.context\":\"$metadata#apis\",\"value\":[{\"labels\":{\"country\":[\"DE\",\"US\"]}}]}";

        JsonNode jsonTree = mapper.readTree(content);
        aggregator.aggregate(jsonTree);
        String actualContent = mapper.writeValueAsString(jsonTree);

        assertEquals(expectedContent, actualContent);
    }
}
