/*
 * Copyright (C) 2018 Onchain <onchain@onchain.com>
 *
 * This file is part of The ontology_Zero.
 *
 * The ontology_Zero is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology_Zero is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology_Zero.  If not, see <http://www.gnu.org/licenses/>.
 */

package neovm

import "github.com/Ontology/vm/neovm/types"

type StackItem struct {
	_object types.StackItemInterface
}

func NewStackItem(object types.StackItemInterface) *StackItem {
	var stackItem StackItem
	stackItem._object = object
	return &stackItem
}

func (s *StackItem) GetStackItem() types.StackItemInterface {
	return s._object
}

func (s *StackItem) GetExecutionContext() *ExecutionContext {
	return nil
}
