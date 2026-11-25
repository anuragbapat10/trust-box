# PROJECT RULES

- **Project Name** - Trustbox
- **Tech Stack** - Go
- **Structure** - Monorepo. Three services planned: a client facing API (`cmd/api` — not yet created), a control plane with datastore access (`cmd/controlplane`), and a node daemon running on gRPC to orchestrate the actual instances locally (`cmd/daemon`). They are three separate packages so that they can be deployed as individual microservices. The directory cmd/ contains each service. Code inside internal/ contains common code.  
- Refer to README.md for additional project context
- If PERSONAL.md exists alongside this file, follow it as additional developer-local instructions (gitignored, never pushed).
