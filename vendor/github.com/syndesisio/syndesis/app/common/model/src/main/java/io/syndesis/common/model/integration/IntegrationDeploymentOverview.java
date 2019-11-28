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
package io.syndesis.common.model.integration;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import org.immutables.value.Value;

@Value.Immutable
@JsonDeserialize(builder = IntegrationDeploymentOverview.Builder.class)
@SuppressWarnings("immutables")
public interface IntegrationDeploymentOverview extends IntegrationDeploymentBase {

    // allow access to ImmutableIntegrationDeploymentOverview.Builder
    class Builder extends ImmutableIntegrationDeploymentOverview.Builder {
    }

    static IntegrationDeploymentOverview of(IntegrationDeployment deployment) {
        return new IntegrationDeploymentOverview.Builder().createFrom(deployment).build();
    }
}
