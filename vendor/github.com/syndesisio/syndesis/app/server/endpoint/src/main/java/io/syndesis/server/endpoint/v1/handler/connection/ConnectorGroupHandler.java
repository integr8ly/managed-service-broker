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
package io.syndesis.server.endpoint.v1.handler.connection;

import javax.ws.rs.Path;

import io.syndesis.server.dao.manager.DataManager;
import io.syndesis.common.model.Kind;
import io.syndesis.common.model.connection.ConnectorGroup;
import io.syndesis.server.endpoint.v1.handler.BaseHandler;
import io.syndesis.server.endpoint.v1.operations.Getter;
import io.syndesis.server.endpoint.v1.operations.Lister;
import io.swagger.annotations.Api;

import org.springframework.stereotype.Component;

@Path("/connectorgroups")
@Api(value = "connectorgroups")
@Component
public class ConnectorGroupHandler extends BaseHandler implements Lister<ConnectorGroup>, Getter<ConnectorGroup> {

    public ConnectorGroupHandler(DataManager dataMgr) {
        super(dataMgr);
    }

    @Override
    public Kind resourceKind() {
        return Kind.ConnectorGroup;
    }

}
