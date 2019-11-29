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

import io.syndesis.integration.component.proxy.ComponentProxyComponent;
import org.apache.camel.test.junit4.CamelTestSupport;

import java.util.UUID;

public abstract class AbstractGoogleSheetsCustomizerTestSupport extends CamelTestSupport {

    private ComponentProxyComponent component = new ComponentProxyComponent("google-sheets-1", "google-sheets");
    private String spreadsheetId = UUID.randomUUID().toString();

    @Override
    public boolean isUseRouteBuilder() {
        return false;
    }

    @Override
    public boolean isCreateCamelContextPerClass() {
        // only create the context once for this class
        return true;
    }

    /**
     * Gets the test component that is about to be customized.
     *
     * @return
     */
    public ComponentProxyComponent getComponent() {
        return component;
    }

    /**
     * Gets the test spreadsheetId.
     *
     * @return
     */
    public String getSpreadsheetId() {
        return spreadsheetId;
    }
}
