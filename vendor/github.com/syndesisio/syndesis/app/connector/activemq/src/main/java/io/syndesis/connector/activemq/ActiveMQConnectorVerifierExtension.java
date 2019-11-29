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
package io.syndesis.connector.activemq;

import java.util.HashMap;
import java.util.Map;
import javax.jms.Connection;
import javax.jms.ConnectionFactory;
import javax.jms.JMSException;

import org.apache.camel.CamelContext;
import org.apache.camel.component.extension.verifier.DefaultComponentVerifierExtension;
import org.apache.camel.component.extension.verifier.ResultBuilder;
import org.apache.camel.component.extension.verifier.ResultErrorBuilder;
import org.apache.camel.component.extension.verifier.ResultErrorHelper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import io.syndesis.connector.support.util.ConnectorOptions;

public class ActiveMQConnectorVerifierExtension extends DefaultComponentVerifierExtension {
    private static final Logger LOG = LoggerFactory.getLogger(ActiveMQConnectorVerifierExtension.class);

    public ActiveMQConnectorVerifierExtension(CamelContext camelContext) {
        super("activemq-connector", camelContext);
    }

    // *********************************
    // Parameters validation
    // *********************************

    @Override
    protected Result verifyParameters(Map<String, Object> parameters) {
        ResultBuilder builder = ResultBuilder.withStatusAndScope(Result.Status.OK, Scope.PARAMETERS)
            .error(ResultErrorHelper.requiresOption("brokerUrl", parameters));

        if (builder.build().getErrors().isEmpty()) {
            verifyCredentials(builder, parameters);
        }

        return builder.build();
    }

    // *********************************
    // Connectivity validation
    // *********************************

    @Override
    protected Result verifyConnectivity(Map<String, Object> parameters) {
        return ResultBuilder.withStatusAndScope(Result.Status.OK, Scope.CONNECTIVITY)
            .error(parameters, this::verifyCredentials)
            .build();
    }

    private void verifyCredentials(ResultBuilder builder, Map<String, Object> parameters) {
        final String brokerUrl = ConnectorOptions.extractOption(parameters, "brokerUrl");
        final String username = ConnectorOptions.extractOption(parameters, "username");
        final String password = ConnectorOptions.extractOption(parameters, "password");
        final boolean skipCertificateCheck = ConnectorOptions.extractOptionAndMap(parameters,
            "skipCertificateCheck", Boolean::parseBoolean, false);
        final String brokerCertificate = ConnectorOptions.extractOption(parameters, "brokerCertificate");
        final String clientCertificate = ConnectorOptions.extractOption(parameters, "clientCertificate");

        LOG.debug("Validating AMQ connection to {}", brokerUrl);

        Connection connection = null;
        try {
            ConnectionFactory connectionFactory = ActiveMQUtil.createActiveMQConnectionFactory(
                brokerUrl,
                username,
                password,
                brokerCertificate,
                clientCertificate,
                skipCertificateCheck);

            // try to create and start the JMS connection
            connection = connectionFactory.createConnection();
            connection.start();
        } catch (JMSException e) {
            final Map<String, Object> redacted = new HashMap<>(parameters);
            redacted.replace("password", "********");

            LOG.warn("Unable to connect to ActiveMQ Broker with parameters {}, Message: {}, error code: {}",
                redacted, e.getMessage(), e.getErrorCode(), e);

            builder.error(ResultErrorBuilder.withCodeAndDescription(VerificationError.StandardCode.ILLEGAL_PARAMETER_VALUE, e.getMessage())
                .parameterKey("brokerUrl")
                .parameterKey("username")
                .parameterKey("password")
                .build());
        } finally {
            if (connection != null) {
                try {
                    connection.close();
                } catch (JMSException ignored) {
                    // ignore close errors
                }
            }
        }
    }

}
