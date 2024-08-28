/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"
export type ExpiresAfter = {
  anchor?: string
  days?: string
}

export type VectorStoreFileCounts = {
  inProgress?: string
  completed?: string
  failed?: string
  cancelled?: string
  total?: string
}

export type VectorStore = {
  id?: string
  object?: string
  createdAt?: string
  name?: string
  usageBytes?: string
  fileCounts?: VectorStoreFileCounts
  status?: string
  expiresAfter?: ExpiresAfter
  expiresAt?: string
  lastActiveAt?: string
  metadata?: {[key: string]: string}
}

export type ChunkingStrategyStatic = {
  maxChunkSizeTokens?: string
  chunkOverlapTokens?: string
}

export type ChunkingStrategy = {
  type?: string
  static?: ChunkingStrategyStatic
}

export type CreateVectorStoreRequest = {
  fileIds?: string[]
  name?: string
  expiresAfter?: ExpiresAfter
  chunkingStrategy?: ChunkingStrategy
  metadata?: {[key: string]: string}
}

export type ListVectorStoresRequest = {
  limit?: number
  order?: string
  after?: string
  berfore?: string
}

export type ListVectorStoresResponse = {
  object?: string
  data?: VectorStore[]
  firstId?: string
  lastId?: string
  hasMore?: boolean
}

export type GetVectorStoreRequest = {
  id?: string
}

export type GetVectorStoreByNameRequest = {
  name?: string
}

export type UpdateVectorStoreRequest = {
  id?: string
  name?: string
  expiresAfter?: ExpiresAfter
  metadata?: {[key: string]: string}
}

export type DeleteVectorStoreRequest = {
  id?: string
}

export type DeleteVectorStoreResponse = {
  id?: string
  object?: string
  deleted?: boolean
}

export type VectorStoreFileError = {
  code?: string
  message?: string
}

export type VectorStoreFile = {
  id?: string
  object?: string
  usageBytes?: string
  createdAt?: string
  vectorStoreId?: string
  status?: string
  lastError?: VectorStoreFileError
  chunkingStrategy?: ChunkingStrategy
}

export type CreateVectorStoreFileRequest = {
  vectorStoreId?: string
  fileId?: string
  chunkingStrategy?: ChunkingStrategy
}

export type ListVectorStoreFilesRequest = {
  vectorStoreId?: string
  limit?: number
  order?: string
  after?: string
  berfore?: string
  filter?: string
}

export type ListVectorStoreFilesResponse = {
  object?: string
  data?: VectorStoreFile[]
  firstId?: string
  lastId?: string
  hasMore?: boolean
}

export type GetVectorStoreFileRequest = {
  vectorStoreId?: string
  fileId?: string
}

export type DeleteVectorStoreFileRequest = {
  vectorStoreId?: string
  fileId?: string
}

export type DeleteVectorStoreFileResponse = {
  id?: string
  object?: string
  deleted?: boolean
}

export type SearchVectorStoreRequest = {
  vectorStoreId?: string
  query?: string
  numDocuments?: number
}

export type SearchVectorStoreResponse = {
  documents?: string[]
}

export class VectorStoreService {
  static CreateVectorStore(req: CreateVectorStoreRequest, initReq?: fm.InitReq): Promise<VectorStore> {
    return fm.fetchReq<CreateVectorStoreRequest, VectorStore>(`/v1/vector_stores`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListVectorStores(req: ListVectorStoresRequest, initReq?: fm.InitReq): Promise<ListVectorStoresResponse> {
    return fm.fetchReq<ListVectorStoresRequest, ListVectorStoresResponse>(`/v1/vector_stores?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetVectorStore(req: GetVectorStoreRequest, initReq?: fm.InitReq): Promise<VectorStore> {
    return fm.fetchReq<GetVectorStoreRequest, VectorStore>(`/v1/vector_stores/${req["id"]}?${fm.renderURLSearchParams(req, ["id"])}`, {...initReq, method: "GET"})
  }
  static GetVectorStoreByName(req: GetVectorStoreByNameRequest, initReq?: fm.InitReq): Promise<VectorStore> {
    return fm.fetchReq<GetVectorStoreByNameRequest, VectorStore>(`/llmoperator.vector_store.v1.VectorStoreService/GetVectorStoreByName`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static UpdateVectorStore(req: UpdateVectorStoreRequest, initReq?: fm.InitReq): Promise<VectorStore> {
    return fm.fetchReq<UpdateVectorStoreRequest, VectorStore>(`/v1/vector_stores/${req["id"]}`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static DeleteVectorStore(req: DeleteVectorStoreRequest, initReq?: fm.InitReq): Promise<DeleteVectorStoreResponse> {
    return fm.fetchReq<DeleteVectorStoreRequest, DeleteVectorStoreResponse>(`/v1/vector_stores/${req["id"]}`, {...initReq, method: "DELETE"})
  }
  static CreateVectorStoreFile(req: CreateVectorStoreFileRequest, initReq?: fm.InitReq): Promise<VectorStoreFile> {
    return fm.fetchReq<CreateVectorStoreFileRequest, VectorStoreFile>(`/v1/vector_stores/${req["vectorStoreId"]}/files`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListVectorStoreFiles(req: ListVectorStoreFilesRequest, initReq?: fm.InitReq): Promise<ListVectorStoreFilesResponse> {
    return fm.fetchReq<ListVectorStoreFilesRequest, ListVectorStoreFilesResponse>(`/v1/vector_stores/${req["vectorStoreId"]}/files?${fm.renderURLSearchParams(req, ["vectorStoreId"])}`, {...initReq, method: "GET"})
  }
  static GetVectorStoreFile(req: GetVectorStoreFileRequest, initReq?: fm.InitReq): Promise<VectorStoreFile> {
    return fm.fetchReq<GetVectorStoreFileRequest, VectorStoreFile>(`/v1/vector_stores/${req["vectorStoreId"]}/files/${req["fileId"]}?${fm.renderURLSearchParams(req, ["vectorStoreId", "fileId"])}`, {...initReq, method: "GET"})
  }
  static DeleteVectorStoreFile(req: DeleteVectorStoreFileRequest, initReq?: fm.InitReq): Promise<DeleteVectorStoreFileResponse> {
    return fm.fetchReq<DeleteVectorStoreFileRequest, DeleteVectorStoreFileResponse>(`/v1/vector_stores/${req["vectorStoreId"]}/files/${req["fileId"]}`, {...initReq, method: "DELETE"})
  }
}
export class VectorStoreInternalService {
  static SearchVectorStore(req: SearchVectorStoreRequest, initReq?: fm.InitReq): Promise<SearchVectorStoreResponse> {
    return fm.fetchReq<SearchVectorStoreRequest, SearchVectorStoreResponse>(`/llmoperator.vector_store.v1.VectorStoreInternalService/SearchVectorStore`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}