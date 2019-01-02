package fly

import (
	"strings"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/util"
)

// GCPPipeline is GCP specific implementation of Pipeline interface
type GCPPipeline struct {
	GCPDefaultRegion     string
	GCPCredsPath         string
	Deployment           string
	FlagGCPRegion        string
	FlagDomain           string
	FlagGithubAuthID     string
	FlagGithubAuthSecret string
	FlagTLSCert          string
	FlagTLSKey           string
	FlagWebSize          string
	FlagWorkerSize       string
	FlagWorkers          int
	ConcourseUpVersion   string
	Namespace            string
}

// NewGCPPipeline return GCPPipeline
func NewGCPPipeline(credsPath string) Pipeline {
	return GCPPipeline{
		GCPCredsPath: credsPath,
	}
}

//BuildPipelineParams builds params for AWS concourse-up self update pipeline
func (a GCPPipeline) BuildPipelineParams(config config.Config) (Pipeline, error) {
	var (
		domain        string
		concourseCert string
		concourseKey  string
	)

	if !validIP4(config.Domain) {
		domain = config.Domain
	}

	if domain != "" {
		concourseCert = config.ConcourseCert
		concourseKey = config.ConcourseKey
	}

	return GCPPipeline{
		GCPCredsPath:         a.GCPCredsPath,
		Deployment:           strings.TrimPrefix(config.Deployment, "concourse-up-"),
		FlagGCPRegion:        config.Region,
		FlagDomain:           domain,
		FlagGithubAuthID:     config.GithubClientID,
		FlagGithubAuthSecret: config.GithubClientSecret,
		FlagTLSCert:          concourseCert,
		FlagTLSKey:           concourseKey,
		FlagWebSize:          config.ConcourseWebSize,
		FlagWorkerSize:       config.ConcourseWorkerSize,
		FlagWorkers:          config.ConcourseWorkerCount,
		ConcourseUpVersion:   ConcourseUpVersion,
		Namespace:            config.Namespace,
	}, nil
}

// GetConfigTemplate returns template for AWS Concourse Up self update pipeline
func (a GCPPipeline) GetConfigTemplate() string {
	return gcpPipelineTemplate

}

// Indent is a helper function to indent the field a given number of spaces
func (a GCPPipeline) Indent(countStr, field string) string {
	return util.Indent(countStr, field)
}

const gcpPipelineTemplate = `
---
resources:
- name: concourse-up-release
  type: github-release
  source:
    user: engineerbetter
    repository: concourse-up
    pre_release: true
- name: every-month
  type: time
  source: {interval: 730h}

jobs:
- name: self-update
  serial_groups: [cup]
  serial: true
  plan:
  - get: concourse-up-release
    trigger: true
  - task: update
    params:
      AWS_REGION: "{{ .FlagGCPRegion }}"
      DOMAIN: "{{ .FlagDomain }}"
      TLS_CERT: |-
        {{ .Indent "8" .FlagTLSCert }}
      TLS_KEY: |-
        {{ .Indent "8" .FlagTLSKey }}
      WORKERS: "{{ .FlagWorkers }}"
      WORKER_SIZE: "{{ .FlagWorkerSize }}"
      WEB_SIZE: "{{ .FlagWebSize }}"
      DEPLOYMENT: "{{ .Deployment }}"
      GITHUB_AUTH_CLIENT_ID: "{{ .FlagGithubAuthID }}"
      GITHUB_AUTH_CLIENT_SECRET: "{{ .FlagGithubAuthSecret }}"
      GOOGLE_APPLICATION_CREDENTIALS: "{{ .GCPCredsPath }}"
      IAAS: GCP
      SELF_UPDATE: true
      NAMESPACE: {{ .Namespace }}
    config:
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: engineerbetter/pcf-ops
      inputs:
      - name: concourse-up-release
      run:
        path: bash
        args:
        - -c
        - |
          set -eux

          cd concourse-up-release
          chmod +x concourse-up-linux-amd64
          ./concourse-up-linux-amd64 deploy $DEPLOYMENT
- name: renew-cert
  serial_groups: [cup]
  serial: true
  plan:
  - get: concourse-up-release
    version: {tag: {{ .ConcourseUpVersion }} }
  - get: every-month
    trigger: true
  - task: update
    params:
      AWS_REGION: "{{ .FlagGCPRegion }}"
      DOMAIN: "{{ .FlagDomain }}"
      TLS_CERT: |-
        {{ .Indent "8" .FlagTLSCert }}
      TLS_KEY: |-
        {{ .Indent "8" .FlagTLSKey }}
      WORKERS: "{{ .FlagWorkers }}"
      WORKER_SIZE: "{{ .FlagWorkerSize }}"
      WEB_SIZE: "{{ .FlagWebSize }}"
      DEPLOYMENT: "{{ .Deployment }}"
      GITHUB_AUTH_CLIENT_ID: "{{ .FlagGithubAuthID }}"
      GITHUB_AUTH_CLIENT_SECRET: "{{ .FlagGithubAuthSecret }}"
      GOOGLE_APPLICATION_CREDENTIALS: "{{ .GCPCredsPath }}"
      IAAS: GCP
      SELF_UPDATE: true
      NAMESPACE: {{ .Namespace }}
    config:
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: engineerbetter/pcf-ops
      inputs:
      - name: concourse-up-release
      run:
        path: bash
        args:
        - -c
        - |
          set -eux

          cd concourse-up-release
          chmod +x concourse-up-linux-amd64
          ./concourse-up-linux-amd64 deploy $DEPLOYMENT
`