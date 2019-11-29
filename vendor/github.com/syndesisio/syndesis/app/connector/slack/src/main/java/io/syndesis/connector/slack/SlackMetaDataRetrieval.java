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
package io.syndesis.connector.slack;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Set;

import org.apache.camel.CamelContext;
import org.apache.camel.component.extension.MetaDataExtension;

import io.syndesis.connector.support.verifier.api.ComponentMetadataRetrieval;
import io.syndesis.connector.support.verifier.api.PropertyPair;
import io.syndesis.connector.support.verifier.api.SyndesisMetadata;

public class SlackMetaDataRetrieval extends ComponentMetadataRetrieval {

    /**
     * TODO: use local extension, remove when switching to camel 2.22.x
     */
    @Override
    protected MetaDataExtension resolveMetaDataExtension(CamelContext context, Class<? extends MetaDataExtension> metaDataExtensionClass, String componentId, String actionId) {
        return new SlackMetaDataExtension(context);
    }

    @SuppressWarnings("unchecked")
    @Override
    protected SyndesisMetadata adapt(CamelContext context, String componentId, String actionId, Map<String, Object> properties, MetaDataExtension.MetaData metadata) {
        try {
            Set<String> channels = (Set<String>) metadata.getPayload();

            List<PropertyPair> channelsResult = new ArrayList<>();
            channels.stream().forEach(
                t -> channelsResult.add(new PropertyPair(t, t))
            );

            return SyndesisMetadata.of(
                Collections.singletonMap("channel", channelsResult)
            );
        } catch ( Exception e) {
            return SyndesisMetadata.EMPTY;
        }
    }

}
