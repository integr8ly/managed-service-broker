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
package io.syndesis.common.model.connection;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import io.syndesis.common.model.Kind;
import io.syndesis.common.model.WithId;
import io.syndesis.common.model.bulletin.ConnectionBulletinBoard;
import io.syndesis.common.util.IndexedProperty;
import org.immutables.value.Value;

@Value.Immutable
@JsonDeserialize(builder = ConnectionOverview.Builder.class)
@SuppressWarnings("immutables")
@IndexedProperty("connectorId")
public interface ConnectionOverview extends WithId<ConnectionOverview>, ConnectionBase {
    @Override
    default Kind getKind() {
        return Kind.ConnectionOverview;
    }

    @Value.Default
    default ConnectionBulletinBoard getBoard() {
        return ConnectionBulletinBoard.emptyBoard();
    }

    // allow access to ImmutableIntegration.Builder
    class Builder extends ImmutableConnectionOverview.Builder {
    }
}
