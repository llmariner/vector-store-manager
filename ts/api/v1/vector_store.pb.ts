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
  in_progress?: string
  completed?: string
  failed?: string
  cancelled?: string
  total?: string
}

export type VectorStore = {
  id?: string
  object?: string
  created_at?: string
  name?: string
  usage_bytes?: string
  file_counts?: VectorStoreFileCounts
  status?: string
  expires_after?: ExpiresAfter
  expires_at?: string
  last_active_at?: string
  metadata?: {[key: string]: string}
}

export type ChunkingStrategyStatic = {
  max_chunk_size_tokens?: string
  chunk_overlap_tokens?: string
}

export type ChunkingStrategy = {
  type?: string
  static?: ChunkingStrategyStatic
}

export type CreateVectorStoreRequest = {
  file_ids?: string[]
  name?: string
  expires_after?: ExpiresAfter
  chunking_strategy?: ChunkingStrategy
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
  first_id?: string
  last_id?: string
  has_more?: boolean
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
  expires_after?: ExpiresAfter
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
  usage_bytes?: string
  created_at?: string
  vector_store_id?: string
  status?: string
  last_error?: VectorStoreFileError
  chunking_strategy?: ChunkingStrategy
}

export type CreateVectorStoreFileRequest = {
  vector_store_id?: string
  file_id?: string
  chunking_strategy?: ChunkingStrategy
}

export type ListVectorStoreFilesRequest = {
  vector_store_id?: string
  limit?: number
  order?: string
  after?: string
  berfore?: string
  filter?: string
}

export type ListVectorStoreFilesResponse = {
  object?: string
  data?: VectorStoreFile[]
  first_id?: string
  last_id?: string
  has_more?: boolean
}

export type GetVectorStoreFileRequest = {
  vector_store_id?: string
  file_id?: string
}

export type DeleteVectorStoreFileRequest = {
  vector_store_id?: string
  file_id?: string
}

export type DeleteVectorStoreFileResponse = {
  id?: string
  object?: string
  deleted?: boolean
}

export type SearchVectorStoreRequest = {
  vector_store_id?: string
  query?: string
  num_documents?: number
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
    return fm.fetchReq<GetVectorStoreByNameRequest, VectorStore>(`/llmariner.vector_store.v1.VectorStoreService/GetVectorStoreByName`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static UpdateVectorStore(req: UpdateVectorStoreRequest, initReq?: fm.InitReq): Promise<VectorStore> {
    return fm.fetchReq<UpdateVectorStoreRequest, VectorStore>(`/v1/vector_stores/${req["id"]}`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static DeleteVectorStore(req: DeleteVectorStoreRequest, initReq?: fm.InitReq): Promise<DeleteVectorStoreResponse> {
    return fm.fetchReq<DeleteVectorStoreRequest, DeleteVectorStoreResponse>(`/v1/vector_stores/${req["id"]}`, {...initReq, method: "DELETE"})
  }
  static CreateVectorStoreFile(req: CreateVectorStoreFileRequest, initReq?: fm.InitReq): Promise<VectorStoreFile> {
    return fm.fetchReq<CreateVectorStoreFileRequest, VectorStoreFile>(`/v1/vector_stores/${req["vector_store_id"]}/files`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListVectorStoreFiles(req: ListVectorStoreFilesRequest, initReq?: fm.InitReq): Promise<ListVectorStoreFilesResponse> {
    return fm.fetchReq<ListVectorStoreFilesRequest, ListVectorStoreFilesResponse>(`/v1/vector_stores/${req["vector_store_id"]}/files?${fm.renderURLSearchParams(req, ["vector_store_id"])}`, {...initReq, method: "GET"})
  }
  static GetVectorStoreFile(req: GetVectorStoreFileRequest, initReq?: fm.InitReq): Promise<VectorStoreFile> {
    return fm.fetchReq<GetVectorStoreFileRequest, VectorStoreFile>(`/v1/vector_stores/${req["vector_store_id"]}/files/${req["file_id"]}?${fm.renderURLSearchParams(req, ["vector_store_id", "file_id"])}`, {...initReq, method: "GET"})
  }
  static DeleteVectorStoreFile(req: DeleteVectorStoreFileRequest, initReq?: fm.InitReq): Promise<DeleteVectorStoreFileResponse> {
    return fm.fetchReq<DeleteVectorStoreFileRequest, DeleteVectorStoreFileResponse>(`/v1/vector_stores/${req["vector_store_id"]}/files/${req["file_id"]}`, {...initReq, method: "DELETE"})
  }
}
export class VectorStoreInternalService {
  static SearchVectorStore(req: SearchVectorStoreRequest, initReq?: fm.InitReq): Promise<SearchVectorStoreResponse> {
    return fm.fetchReq<SearchVectorStoreRequest, SearchVectorStoreResponse>(`/llmariner.vector_store.v1.VectorStoreInternalService/SearchVectorStore`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}