package client

import "github.com/cry-inc/bdm/pkg/bdm"

// Limit size of JSON payloads (when reading HTTP responses)
const maxBodySize = bdm.JsonSizeLimit
