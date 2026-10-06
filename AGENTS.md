# PROJECT RULES

- **Project Name** - Trustbox
- **Tech Stack** - Go
- **Structure** - Monorepo. Three services, a client facing API (cmd/api), a control plane with datastore access (cmd/control-plane), and a node daemon running on gRPC to orchestrate the actual instances locally (cmd/node-daemon). Will be three separate packages so that they can be deployed as individual microservices. The directory cmd/ contains each service. Code inside internal/ contains common code.  
- Refer to README.md for additional project context
- If PERSONAL.md exists alongside this file, follow it as additional developer-local instructions (gitignored, never pushed).
