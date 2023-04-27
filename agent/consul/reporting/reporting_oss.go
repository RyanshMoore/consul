// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !consulent
// +build !consulent

package reporting

type EntDeps struct{}

func (rm *ReportingManager) initEnterpriseReporting(entDeps EntDeps) error {
	// no op
	return nil
}

func (rm *ReportingManager) StartReportingAgent() error {
	// no op
	return nil
}

func (rm *ReportingManager) StopReportingAgent() error {
	// no op
	return nil
}
