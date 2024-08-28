import * as fm from "../../fetch.pb";
export type ExpiresAfter = {
    anchor?: string;
    days?: string;
};
export type VectorStoreFileCounts = {
    inProgress?: string;
    completed?: string;
    failed?: string;
    cancelled?: string;
    total?: string;
};
export type VectorStore = {
    id?: string;
    object?: string;
    createdAt?: string;
    name?: string;
    usageBytes?: string;
    fileCounts?: VectorStoreFileCounts;
    status?: string;
    expiresAfter?: ExpiresAfter;
    expiresAt?: string;
    lastActiveAt?: string;
    metadata?: {
        [key: string]: string;
    };
};
export type ChunkingStrategyStatic = {
    maxChunkSizeTokens?: string;
    chunkOverlapTokens?: string;
};
export type ChunkingStrategy = {
    type?: string;
    static?: ChunkingStrategyStatic;
};
export type CreateVectorStoreRequest = {
    fileIds?: string[];
    name?: string;
    expiresAfter?: ExpiresAfter;
    chunkingStrategy?: ChunkingStrategy;
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
    firstId?: string;
    lastId?: string;
    hasMore?: boolean;
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
    expiresAfter?: ExpiresAfter;
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
    usageBytes?: string;
    createdAt?: string;
    vectorStoreId?: string;
    status?: string;
    lastError?: VectorStoreFileError;
    chunkingStrategy?: ChunkingStrategy;
};
export type CreateVectorStoreFileRequest = {
    vectorStoreId?: string;
    fileId?: string;
    chunkingStrategy?: ChunkingStrategy;
};
export type ListVectorStoreFilesRequest = {
    vectorStoreId?: string;
    limit?: number;
    order?: string;
    after?: string;
    berfore?: string;
    filter?: string;
};
export type ListVectorStoreFilesResponse = {
    object?: string;
    data?: VectorStoreFile[];
    firstId?: string;
    lastId?: string;
    hasMore?: boolean;
};
export type GetVectorStoreFileRequest = {
    vectorStoreId?: string;
    fileId?: string;
};
export type DeleteVectorStoreFileRequest = {
    vectorStoreId?: string;
    fileId?: string;
};
export type DeleteVectorStoreFileResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type SearchVectorStoreRequest = {
    vectorStoreId?: string;
    query?: string;
    numDocuments?: number;
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
