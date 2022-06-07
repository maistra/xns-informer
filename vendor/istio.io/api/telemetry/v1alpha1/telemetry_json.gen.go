// Code generated by protoc-gen-jsonshim. DO NOT EDIT.
package v1alpha1

import (
	bytes "bytes"
	jsonpb "github.com/golang/protobuf/jsonpb"
)

// MarshalJSON is a custom marshaler for Telemetry
func (this *Telemetry) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for Telemetry
func (this *Telemetry) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for Tracing
func (this *Tracing) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for Tracing
func (this *Tracing) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for Tracing_TracingSelector
func (this *Tracing_TracingSelector) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for Tracing_TracingSelector
func (this *Tracing_TracingSelector) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for Tracing_CustomTag
func (this *Tracing_CustomTag) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for Tracing_CustomTag
func (this *Tracing_CustomTag) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for Tracing_Literal
func (this *Tracing_Literal) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for Tracing_Literal
func (this *Tracing_Literal) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for Tracing_Environment
func (this *Tracing_Environment) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for Tracing_Environment
func (this *Tracing_Environment) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for Tracing_RequestHeader
func (this *Tracing_RequestHeader) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for Tracing_RequestHeader
func (this *Tracing_RequestHeader) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for ProviderRef
func (this *ProviderRef) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for ProviderRef
func (this *ProviderRef) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for Metrics
func (this *Metrics) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for Metrics
func (this *Metrics) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for MetricSelector
func (this *MetricSelector) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for MetricSelector
func (this *MetricSelector) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for MetricsOverrides
func (this *MetricsOverrides) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for MetricsOverrides
func (this *MetricsOverrides) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for MetricsOverrides_TagOverride
func (this *MetricsOverrides_TagOverride) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for MetricsOverrides_TagOverride
func (this *MetricsOverrides_TagOverride) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for AccessLogging
func (this *AccessLogging) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for AccessLogging
func (this *AccessLogging) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for AccessLogging_LogSelector
func (this *AccessLogging_LogSelector) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for AccessLogging_LogSelector
func (this *AccessLogging_LogSelector) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for AccessLogging_Filter
func (this *AccessLogging_Filter) MarshalJSON() ([]byte, error) {
	str, err := TelemetryMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for AccessLogging_Filter
func (this *AccessLogging_Filter) UnmarshalJSON(b []byte) error {
	return TelemetryUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

var (
	TelemetryMarshaler   = &jsonpb.Marshaler{}
	TelemetryUnmarshaler = &jsonpb.Unmarshaler{AllowUnknownFields: true}
)
