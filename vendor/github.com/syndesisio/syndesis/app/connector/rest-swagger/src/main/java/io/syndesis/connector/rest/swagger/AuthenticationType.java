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
package io.syndesis.connector.rest.swagger;

import java.util.Objects;

public enum AuthenticationType {
    apiKey, basic, none, oauth2;

    public static AuthenticationType fromString(final String value) {
        Objects.requireNonNull(value, "authenticationType");

        final int idx = value.indexOf(':');
        if (idx > 0) {
            return valueOf(value.substring(0, idx));
        }

        return valueOf(value);
    }
}
