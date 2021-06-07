package client

import "github.com/cry-inc/bdm/pkg/bdm"

// User when by client when sending HTTP requests to the server
const apiTokenField = "bdm-api-token"

// Limit size of JSON payloads (when reading HTTP responses)
const maxBodySize = bdm.JsonSizeLimit
