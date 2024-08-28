/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/
import * as fm from "../../fetch.pb";
export class VectorStoreService {
    static CreateVectorStore(req, initReq) {
        return fm.fetchReq(`/v1/vector_stores`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListVectorStores(req, initReq) {
        return fm.fetchReq(`/v1/vector_stores?${fm.renderURLSearchParams(req, [])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
    static GetVectorStore(req, initReq) {
        return fm.fetchReq(`/v1/vector_stores/${req["id"]}?${fm.renderURLSearchParams(req, ["id"])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
    static GetVectorStoreByName(req, initReq) {
        return fm.fetchReq(`/llmoperator.vector_store.v1.VectorStoreService/GetVectorStoreByName`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static UpdateVectorStore(req, initReq) {
        return fm.fetchReq(`/v1/vector_stores/${req["id"]}`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static DeleteVectorStore(req, initReq) {
        return fm.fetchReq(`/v1/vector_stores/${req["id"]}`, Object.assign(Object.assign({}, initReq), { method: "DELETE" }));
    }
    static CreateVectorStoreFile(req, initReq) {
        return fm.fetchReq(`/v1/vector_stores/${req["vectorStoreId"]}/files`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static ListVectorStoreFiles(req, initReq) {
        return fm.fetchReq(`/v1/vector_stores/${req["vectorStoreId"]}/files?${fm.renderURLSearchParams(req, ["vectorStoreId"])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
    static GetVectorStoreFile(req, initReq) {
        return fm.fetchReq(`/v1/vector_stores/${req["vectorStoreId"]}/files/${req["fileId"]}?${fm.renderURLSearchParams(req, ["vectorStoreId", "fileId"])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
    static DeleteVectorStoreFile(req, initReq) {
        return fm.fetchReq(`/v1/vector_stores/${req["vectorStoreId"]}/files/${req["fileId"]}`, Object.assign(Object.assign({}, initReq), { method: "DELETE" }));
    }
}
export class VectorStoreInternalService {
    static SearchVectorStore(req, initReq) {
        return fm.fetchReq(`/llmoperator.vector_store.v1.VectorStoreInternalService/SearchVectorStore`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
}
