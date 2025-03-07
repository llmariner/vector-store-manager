/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

export enum FieldDescriptorProtoType {
  TYPE_DOUBLE = "TYPE_DOUBLE",
  TYPE_FLOAT = "TYPE_FLOAT",
  TYPE_INT64 = "TYPE_INT64",
  TYPE_UINT64 = "TYPE_UINT64",
  TYPE_INT32 = "TYPE_INT32",
  TYPE_FIXED64 = "TYPE_FIXED64",
  TYPE_FIXED32 = "TYPE_FIXED32",
  TYPE_BOOL = "TYPE_BOOL",
  TYPE_STRING = "TYPE_STRING",
  TYPE_GROUP = "TYPE_GROUP",
  TYPE_MESSAGE = "TYPE_MESSAGE",
  TYPE_BYTES = "TYPE_BYTES",
  TYPE_UINT32 = "TYPE_UINT32",
  TYPE_ENUM = "TYPE_ENUM",
  TYPE_SFIXED32 = "TYPE_SFIXED32",
  TYPE_SFIXED64 = "TYPE_SFIXED64",
  TYPE_SINT32 = "TYPE_SINT32",
  TYPE_SINT64 = "TYPE_SINT64",
}

export enum FieldDescriptorProtoLabel {
  LABEL_OPTIONAL = "LABEL_OPTIONAL",
  LABEL_REQUIRED = "LABEL_REQUIRED",
  LABEL_REPEATED = "LABEL_REPEATED",
}

export enum FileOptionsOptimizeMode {
  SPEED = "SPEED",
  CODE_SIZE = "CODE_SIZE",
  LITE_RUNTIME = "LITE_RUNTIME",
}

export enum FieldOptionsCType {
  STRING = "STRING",
  CORD = "CORD",
  STRING_PIECE = "STRING_PIECE",
}

export enum FieldOptionsJSType {
  JS_NORMAL = "JS_NORMAL",
  JS_STRING = "JS_STRING",
  JS_NUMBER = "JS_NUMBER",
}

export enum MethodOptionsIdempotencyLevel {
  IDEMPOTENCY_UNKNOWN = "IDEMPOTENCY_UNKNOWN",
  NO_SIDE_EFFECTS = "NO_SIDE_EFFECTS",
  IDEMPOTENT = "IDEMPOTENT",
}

export type FileDescriptorSet = {
  file?: FileDescriptorProto[]
}

export type FileDescriptorProto = {
  name?: string
  package?: string
  dependency?: string[]
  public_dependency?: number[]
  weak_dependency?: number[]
  message_type?: DescriptorProto[]
  enum_type?: EnumDescriptorProto[]
  service?: ServiceDescriptorProto[]
  extension?: FieldDescriptorProto[]
  options?: FileOptions
  source_code_info?: SourceCodeInfo
  syntax?: string
}

export type DescriptorProtoExtensionRange = {
  start?: number
  end?: number
  options?: ExtensionRangeOptions
}

export type DescriptorProtoReservedRange = {
  start?: number
  end?: number
}

export type DescriptorProto = {
  name?: string
  field?: FieldDescriptorProto[]
  extension?: FieldDescriptorProto[]
  nested_type?: DescriptorProto[]
  enum_type?: EnumDescriptorProto[]
  extension_range?: DescriptorProtoExtensionRange[]
  oneof_decl?: OneofDescriptorProto[]
  options?: MessageOptions
  reserved_range?: DescriptorProtoReservedRange[]
  reserved_name?: string[]
}

export type ExtensionRangeOptions = {
  uninterpreted_option?: UninterpretedOption[]
}

export type FieldDescriptorProto = {
  name?: string
  number?: number
  label?: FieldDescriptorProtoLabel
  type?: FieldDescriptorProtoType
  type_name?: string
  extendee?: string
  default_value?: string
  oneof_index?: number
  json_name?: string
  options?: FieldOptions
  proto3_optional?: boolean
}

export type OneofDescriptorProto = {
  name?: string
  options?: OneofOptions
}

export type EnumDescriptorProtoEnumReservedRange = {
  start?: number
  end?: number
}

export type EnumDescriptorProto = {
  name?: string
  value?: EnumValueDescriptorProto[]
  options?: EnumOptions
  reserved_range?: EnumDescriptorProtoEnumReservedRange[]
  reserved_name?: string[]
}

export type EnumValueDescriptorProto = {
  name?: string
  number?: number
  options?: EnumValueOptions
}

export type ServiceDescriptorProto = {
  name?: string
  method?: MethodDescriptorProto[]
  options?: ServiceOptions
}

export type MethodDescriptorProto = {
  name?: string
  input_type?: string
  output_type?: string
  options?: MethodOptions
  client_streaming?: boolean
  server_streaming?: boolean
}

export type FileOptions = {
  java_package?: string
  java_outer_classname?: string
  java_multiple_files?: boolean
  java_generate_equals_and_hash?: boolean
  java_string_check_utf8?: boolean
  optimize_for?: FileOptionsOptimizeMode
  go_package?: string
  cc_generic_services?: boolean
  java_generic_services?: boolean
  py_generic_services?: boolean
  php_generic_services?: boolean
  deprecated?: boolean
  cc_enable_arenas?: boolean
  objc_class_prefix?: string
  csharp_namespace?: string
  swift_prefix?: string
  php_class_prefix?: string
  php_namespace?: string
  php_metadata_namespace?: string
  ruby_package?: string
  uninterpreted_option?: UninterpretedOption[]
}

export type MessageOptions = {
  message_set_wire_format?: boolean
  no_standard_descriptor_accessor?: boolean
  deprecated?: boolean
  map_entry?: boolean
  uninterpreted_option?: UninterpretedOption[]
}

export type FieldOptions = {
  ctype?: FieldOptionsCType
  packed?: boolean
  jstype?: FieldOptionsJSType
  lazy?: boolean
  unverified_lazy?: boolean
  deprecated?: boolean
  weak?: boolean
  uninterpreted_option?: UninterpretedOption[]
}

export type OneofOptions = {
  uninterpreted_option?: UninterpretedOption[]
}

export type EnumOptions = {
  allow_alias?: boolean
  deprecated?: boolean
  uninterpreted_option?: UninterpretedOption[]
}

export type EnumValueOptions = {
  deprecated?: boolean
  uninterpreted_option?: UninterpretedOption[]
}

export type ServiceOptions = {
  deprecated?: boolean
  uninterpreted_option?: UninterpretedOption[]
}

export type MethodOptions = {
  deprecated?: boolean
  idempotency_level?: MethodOptionsIdempotencyLevel
  uninterpreted_option?: UninterpretedOption[]
}

export type UninterpretedOptionNamePart = {
  name_part?: string
  is_extension?: boolean
}

export type UninterpretedOption = {
  name?: UninterpretedOptionNamePart[]
  identifier_value?: string
  positive_int_value?: string
  negative_int_value?: string
  double_value?: number
  string_value?: Uint8Array
  aggregate_value?: string
}

export type SourceCodeInfoLocation = {
  path?: number[]
  span?: number[]
  leading_comments?: string
  trailing_comments?: string
  leading_detached_comments?: string[]
}

export type SourceCodeInfo = {
  location?: SourceCodeInfoLocation[]
}

export type GeneratedCodeInfoAnnotation = {
  path?: number[]
  source_file?: string
  begin?: number
  end?: number
}

export type GeneratedCodeInfo = {
  annotation?: GeneratedCodeInfoAnnotation[]
}