/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../../fetch.pb"
import * as LlmarinerVector_storeV1Vector_store from "../vector_store.pb"
export class VectorStoreService {
  static GetVectorStoreByName(req: LlmarinerVector_storeV1Vector_store.GetVectorStoreByNameRequest, initReq?: fm.InitReq): Promise<LlmarinerVector_storeV1Vector_store.VectorStore> {
    return fm.fetchReq<LlmarinerVector_storeV1Vector_store.GetVectorStoreByNameRequest, LlmarinerVector_storeV1Vector_store.VectorStore>(`/llmoperator.vector_store.v1.VectorStoreService/GetVectorStoreByName`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}
export class VectorStoreInternalService {
  static SearchVectorStore(req: LlmarinerVector_storeV1Vector_store.SearchVectorStoreRequest, initReq?: fm.InitReq): Promise<LlmarinerVector_storeV1Vector_store.SearchVectorStoreResponse> {
    return fm.fetchReq<LlmarinerVector_storeV1Vector_store.SearchVectorStoreRequest, LlmarinerVector_storeV1Vector_store.SearchVectorStoreResponse>(`/llmoperator.vector_store.v1.VectorStoreInternalService/SearchVectorStore`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}