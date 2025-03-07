import * as fm from "../../fetch.pb";
export type ExpiresAfter = {
    anchor?: string;
    days?: string;
};
export type VectorStoreFileCounts = {
    in_progress?: string;
    completed?: string;
    failed?: string;
    cancelled?: string;
    total?: string;
};
export type VectorStore = {
    id?: string;
    object?: string;
    created_at?: string;
    name?: string;
    usage_bytes?: string;
    file_counts?: VectorStoreFileCounts;
    status?: string;
    expires_after?: ExpiresAfter;
    expires_at?: string;
    last_active_at?: string;
    metadata?: {
        [key: string]: string;
    };
};
export type ChunkingStrategyStatic = {
    max_chunk_size_tokens?: string;
    chunk_overlap_tokens?: string;
};
export type ChunkingStrategy = {
    type?: string;
    static?: ChunkingStrategyStatic;
};
export type CreateVectorStoreRequest = {
    file_ids?: string[];
    name?: string;
    expires_after?: ExpiresAfter;
    chunking_strategy?: ChunkingStrategy;
    metadata?: {
        [key: string]: string;
    };
};
export type ListVectorStoresRequest = {
    limit?: number;
    order?: string;
    after?: string;
    berfore?: string;
};
export type ListVectorStoresResponse = {
    object?: string;
    data?: VectorStore[];
    first_id?: string;
    last_id?: string;
    has_more?: boolean;
};
export type GetVectorStoreRequest = {
    id?: string;
};
export type GetVectorStoreByNameRequest = {
    name?: string;
};
export type UpdateVectorStoreRequest = {
    id?: string;
    name?: string;
    expires_after?: ExpiresAfter;
    metadata?: {
        [key: string]: string;
    };
};
export type DeleteVectorStoreRequest = {
    id?: string;
};
export type DeleteVectorStoreResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type VectorStoreFileError = {
    code?: string;
    message?: string;
};
export type VectorStoreFile = {
    id?: string;
    object?: string;
    usage_bytes?: string;
    created_at?: string;
    vector_store_id?: string;
    status?: string;
    last_error?: VectorStoreFileError;
    chunking_strategy?: ChunkingStrategy;
};
export type CreateVectorStoreFileRequest = {
    vector_store_id?: string;
    file_id?: string;
    chunking_strategy?: ChunkingStrategy;
};
export type ListVectorStoreFilesRequest = {
    vector_store_id?: string;
    limit?: number;
    order?: string;
    after?: string;
    berfore?: string;
    filter?: string;
};
export type ListVectorStoreFilesResponse = {
    object?: string;
    data?: VectorStoreFile[];
    first_id?: string;
    last_id?: string;
    has_more?: boolean;
};
export type GetVectorStoreFileRequest = {
    vector_store_id?: string;
    file_id?: string;
};
export type DeleteVectorStoreFileRequest = {
    vector_store_id?: string;
    file_id?: string;
};
export type DeleteVectorStoreFileResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type SearchVectorStoreRequest = {
    vector_store_id?: string;
    query?: string;
    num_documents?: number;
};
export type SearchVectorStoreResponse = {
    documents?: string[];
};
export declare class VectorStoreService {
    static CreateVectorStore(req: CreateVectorStoreRequest, initReq?: fm.InitReq): Promise<VectorStore>;
    static ListVectorStores(req: ListVectorStoresRequest, initReq?: fm.InitReq): Promise<ListVectorStoresResponse>;
    static GetVectorStore(req: GetVectorStoreRequest, initReq?: fm.InitReq): Promise<VectorStore>;
    static GetVectorStoreByName(req: GetVectorStoreByNameRequest, initReq?: fm.InitReq): Promise<VectorStore>;
    static UpdateVectorStore(req: UpdateVectorStoreRequest, initReq?: fm.InitReq): Promise<VectorStore>;
    static DeleteVectorStore(req: DeleteVectorStoreRequest, initReq?: fm.InitReq): Promise<DeleteVectorStoreResponse>;
    static CreateVectorStoreFile(req: CreateVectorStoreFileRequest, initReq?: fm.InitReq): Promise<VectorStoreFile>;
    static ListVectorStoreFiles(req: ListVectorStoreFilesRequest, initReq?: fm.InitReq): Promise<ListVectorStoreFilesResponse>;
    static GetVectorStoreFile(req: GetVectorStoreFileRequest, initReq?: fm.InitReq): Promise<VectorStoreFile>;
    static DeleteVectorStoreFile(req: DeleteVectorStoreFileRequest, initReq?: fm.InitReq): Promise<DeleteVectorStoreFileResponse>;
}
export declare class VectorStoreInternalService {
    static SearchVectorStore(req: SearchVectorStoreRequest, initReq?: fm.InitReq): Promise<SearchVectorStoreResponse>;
}
