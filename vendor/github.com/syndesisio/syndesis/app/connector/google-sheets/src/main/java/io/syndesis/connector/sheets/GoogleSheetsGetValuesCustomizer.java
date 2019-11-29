/*
 * Copyright (C) 2016 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package io.syndesis.connector.sheets;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import org.apache.camel.Exchange;
import org.apache.camel.Message;
import org.apache.camel.component.google.sheets.internal.GoogleSheetsApiCollection;
import org.apache.camel.component.google.sheets.internal.SheetsSpreadsheetsValuesApiMethod;
import org.apache.camel.component.google.sheets.stream.GoogleSheetsStreamConstants;
import org.apache.camel.util.ObjectHelper;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.google.api.services.sheets.v4.model.ValueRange;
import io.syndesis.common.util.Json;
import io.syndesis.connector.sheets.model.CellCoordinate;
import io.syndesis.connector.sheets.model.RangeCoordinate;
import io.syndesis.connector.support.util.ConnectorOptions;
import io.syndesis.integration.component.proxy.ComponentProxyComponent;
import io.syndesis.integration.component.proxy.ComponentProxyCustomizer;

public class GoogleSheetsGetValuesCustomizer implements ComponentProxyCustomizer {

    private static final String ROW_PREFIX = "#";

    private String spreadsheetId;
    private String range;
    private String majorDimension;
    private String[] columnNames;
    private boolean splitResults;

    @Override
    public void customize(ComponentProxyComponent component, Map<String, Object> options) {
        setApiMethod(options);
        component.setBeforeConsumer(this::beforeConsumer);
    }

    private void setApiMethod(Map<String, Object> options) {
        spreadsheetId = ConnectorOptions.extractOption(options, "spreadsheetId");
        range = ConnectorOptions.extractOption(options, "range");
        majorDimension = ConnectorOptions.extractOption(options, "majorDimension", RangeCoordinate.DIMENSION_ROWS);
        columnNames = ConnectorOptions.extractOptionAndMap(options, "columnNames",
            names -> names.split(","), new String[]{});
        Arrays.parallelSetAll(columnNames, (i) -> columnNames[i].trim());

        splitResults = ConnectorOptions.extractOptionAndMap(options, "splitResults", Boolean::valueOf, false);

        options.put("apiName",
                GoogleSheetsApiCollection.getCollection().getApiName(SheetsSpreadsheetsValuesApiMethod.class).getName());
        options.put("methodName", "get");
    }

    private void beforeConsumer(Exchange exchange) throws JsonProcessingException {
        final Message in = exchange.getIn();

        if (splitResults) {
            in.setBody(createModelFromSplitValues(in));
        } else {
            in.setBody(createModelFromValueRange(in));
        }
    }

    private List<String> createModelFromValueRange(Message in) throws JsonProcessingException {
        final List<String> jsonBeans = new ArrayList<>();
        final ValueRange valueRange = in.getBody(ValueRange.class);

        if (valueRange != null) {
            if (ObjectHelper.isNotEmpty(valueRange.getRange())) {
                range = valueRange.getRange();
            }
            RangeCoordinate rangeCoordinate = RangeCoordinate.fromRange(range);

            if (ObjectHelper.isNotEmpty(valueRange.getMajorDimension())) {
                majorDimension = valueRange.getMajorDimension();
            }

            if (ObjectHelper.equal(RangeCoordinate.DIMENSION_ROWS, majorDimension)) {
                for (List<Object> values : valueRange.getValues()) {
                    final Map<String, Object> model = new HashMap<>();
                    model.put("spreadsheetId", spreadsheetId);
                    int columnIndex = rangeCoordinate.getColumnStartIndex();
                    for (Object value : values) {
                        model.put(CellCoordinate.getColumnName(columnIndex, rangeCoordinate.getColumnStartIndex(), columnNames), value);
                        columnIndex++;
                    }
                    jsonBeans.add(Json.writer().writeValueAsString(model));
                }
            } else if (ObjectHelper.equal(RangeCoordinate.DIMENSION_COLUMNS, majorDimension)) {
                for (List<Object> values : valueRange.getValues()) {
                    final Map<String, Object> model = new HashMap<>();
                    model.put("spreadsheetId", spreadsheetId);
                    int rowIndex = rangeCoordinate.getRowStartIndex() + 1;
                    for (Object value : values) {
                        model.put(ROW_PREFIX + rowIndex, value);
                        rowIndex++;
                    }
                    jsonBeans.add(Json.writer().writeValueAsString(model));
                }
            }
        }

        return jsonBeans;
    }

    private String createModelFromSplitValues(Message in) throws JsonProcessingException {
        final List<?> values = in.getBody(List.class);

        final Map<String, Object> model = new HashMap<>();
        model.put("spreadsheetId", spreadsheetId);

        if (values != null) {
            if (ObjectHelper.isNotEmpty(in.getHeader(GoogleSheetsStreamConstants.RANGE))) {
                range = in.getHeader(GoogleSheetsStreamConstants.RANGE).toString();
            }
            RangeCoordinate rangeCoordinate = RangeCoordinate.fromRange(range);

            if (ObjectHelper.isNotEmpty(in.getHeader(GoogleSheetsStreamConstants.MAJOR_DIMENSION))) {
                majorDimension = in.getHeader(GoogleSheetsStreamConstants.MAJOR_DIMENSION).toString();
            }

            if (ObjectHelper.equal(RangeCoordinate.DIMENSION_ROWS, majorDimension)) {
                int columnIndex = rangeCoordinate.getColumnStartIndex();
                for (Object value : values) {
                    model.put(CellCoordinate.getColumnName(columnIndex, rangeCoordinate.getColumnStartIndex(), columnNames), value);
                    columnIndex++;
                }
            } else if (ObjectHelper.equal(RangeCoordinate.DIMENSION_COLUMNS, majorDimension)) {
                int rowIndex = rangeCoordinate.getRowStartIndex() + 1;
                for (Object value : values) {
                    model.put(ROW_PREFIX + rowIndex, value);
                    rowIndex++;
                }
            }
        }

        return Json.writer().writeValueAsString(model);
    }

}
