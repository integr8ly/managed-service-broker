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
package io.syndesis.common.model;

public interface OutputDataShapeAware {

    /**
     *Sets the output data shape.
     */
    void setOutputDataShape(DataShape dataShape);

    /**
     * Returns the output data shape.
     */
    DataShape getOutputDataShape();

    /**
     * Assign the given dataShape to the target object if it implements
     * InputDataShapeAware interface.
     *
     * @param target the potential InputDataShapeAware target.
     * @param dataShape the data shape.
     */
    static void trySetOutputDataShape(Object target, DataShape dataShape) {
        if (target instanceof OutputDataShapeAware) {
            ((OutputDataShapeAware)target).setOutputDataShape(dataShape);
        }
    }
}
