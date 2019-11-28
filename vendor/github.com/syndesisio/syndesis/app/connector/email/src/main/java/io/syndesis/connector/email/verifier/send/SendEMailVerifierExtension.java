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
package io.syndesis.connector.email.verifier.send;

import java.util.Map;
import javax.mail.Session;
import javax.mail.Transport;
import org.apache.camel.CamelContext;
import org.apache.camel.component.extension.verifier.ResultBuilder;
import org.apache.camel.component.extension.verifier.ResultErrorBuilder;
import org.apache.camel.component.extension.verifier.ResultErrorHelper;
import org.apache.camel.component.mail.JavaMailSender;
import org.apache.camel.component.mail.MailConfiguration;
import org.apache.camel.util.ObjectHelper;
import io.syndesis.connector.email.verifier.AbstractEMailVerifier;
import io.syndesis.connector.support.util.ConnectorOptions;

public class SendEMailVerifierExtension extends AbstractEMailVerifier {

    protected SendEMailVerifierExtension(String defaultScheme, CamelContext context) {
        super(defaultScheme, context);
    }


    // *********************************
    // Parameters validation
    // *********************************

    @Override
    protected Result verifyParameters(Map<String, Object> parameters) {

        ResultBuilder builder = ResultBuilder.withStatusAndScope(Result.Status.OK, Scope.PARAMETERS)
            .error(ResultErrorHelper.requiresOption(HOST, parameters))
            .error(ResultErrorHelper.requiresOption(PORT, parameters))
            .error(ResultErrorHelper.requiresOption(USER, parameters))
            .error(ResultErrorHelper.requiresOption(PASSWORD, parameters));

        //
        // SMTP Protocol is hard-wired as the send-email protocol
        //
        parameters.put(PROTOCOL, Protocol.SMTP.id());

        return builder.build();
    }

    // *********************************
    // Connectivity validation
    // *********************************

    @Override
    @SuppressWarnings("PMD.AvoidCatchingGenericException")
    protected Result verifyConnectivity(Map<String, Object> parameters) {
        ResultBuilder builder = ResultBuilder.withStatusAndScope(Result.Status.OK, Scope.CONNECTIVITY);

        try {
            MailConfiguration configuration = createConfiguration(parameters);

            String timeoutVal = ConnectorOptions.extractOption(parameters, CONNECTION_TIMEOUT);
            if (ObjectHelper.isEmpty(timeoutVal)) {
                timeoutVal = Long.toString(DEFAULT_CONNECTION_TIMEOUT);
            }

            setConnectionTimeoutProperty(parameters, configuration, timeoutVal);

            JavaMailSender sender = createJavaMailSender(configuration);
            Session session = sender.getSession();

            Transport transport = session.getTransport(configuration.getProtocol());
            try {
                transport.connect(configuration.getHost(), configuration.getPort(),
                                  configuration.getUsername(), configuration.getPassword());
            } finally {
                if (transport.isConnected()) {
                    transport.close();
                }
            }
        } catch (Exception e) {
            ResultErrorBuilder errorBuilder = ResultErrorBuilder.withCodeAndDescription(VerificationError.StandardCode.AUTHENTICATION, e.getMessage())
                .detail("mail_exception_message", e.getMessage()).detail(VerificationError.ExceptionAttribute.EXCEPTION_CLASS, e.getClass().getName())
                .detail(VerificationError.ExceptionAttribute.EXCEPTION_INSTANCE, e);

            builder.error(errorBuilder.build());
        }

        return builder.build();
    }
}
