//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the
//  License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an "AS
//  IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language
//  governing permissions and limitations under the License.

package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/couchbaselabs/cbgt"
)

// DeleteIndexHandler is a REST handler that processes an index
// deletion request.
type DeleteIndexHandler struct {
	mgr *cbgt.Manager
}

func NewDeleteIndexHandler(mgr *cbgt.Manager) *DeleteIndexHandler {
	return &DeleteIndexHandler{mgr: mgr}
}

func (h *DeleteIndexHandler) RESTOpts(opts map[string]string) {
	opts["param: indexName"] = "required, string, URL path parameter\n\n" +
		"The name of the index definition to be deleted."
}

func (h *DeleteIndexHandler) ServeHTTP(
	w http.ResponseWriter, req *http.Request) {
	indexName := mux.Vars(req)["indexName"]
	if indexName == "" {
		ShowError(w, req, "rest_delete_index: index name is required", 400)
		return
	}

	err := h.mgr.DeleteIndex(indexName)
	if err != nil {
		ShowError(w, req, fmt.Sprintf("rest_delete_index:"+
			" error deleting index, err: %v", err), 400)
		return
	}

	MustEncode(w, struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	})
}
