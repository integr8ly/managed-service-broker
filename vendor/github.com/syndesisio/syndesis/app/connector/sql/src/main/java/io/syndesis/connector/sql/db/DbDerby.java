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
package io.syndesis.connector.sql.db;

import java.util.Locale;

import org.apache.camel.util.ObjectHelper;

public class DbDerby extends DbStandard {

    @Override
    public String getDefaultSchema(String dbUser) {
        if (ObjectHelper.isNotEmpty(dbUser)) {
            return dbUser.toUpperCase(Locale.US);
        } else {
            return null;
        }
    }

    @Override
    public String getAutoIncrementGrammar() {
        return "INTEGER NOT NULL GENERATED ALWAYS AS IDENTITY (START WITH 1, INCREMENT BY 1)";
    }

    @Override
    public String getName() {
        return "Derby";
    }
}
