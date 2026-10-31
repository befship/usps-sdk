package usps

import (
	"github.com/befship/usps-sdk/internal/generated/addresses"
	"github.com/befship/usps-sdk/internal/generated/labels"
	"github.com/befship/usps-sdk/internal/generated/prices"
	"github.com/befship/usps-sdk/internal/generated/tracking"
)

// Addresses API types.
type (
	AddressesClient        = addresses.ClientWithResponses
	AddressesRequestEditor = addresses.RequestEditorFn
	GetAddressParams       = addresses.GetAddressParams
	AddressResponse        = addresses.AddressResponse
	StandardizedAddress    = addresses.StandardizedAddress
	AddressAdditionalInfo  = addresses.AddressAdditionalInfo
)

// Labels API types.
type (
	LabelsClient               = labels.ClientWithResponses
	LabelsRequestEditor        = labels.RequestEditorFn
	CreateLabelParams          = labels.CreateLabelParams
	GetLabelParams             = labels.GetLabelParams
	CancelLabelParams          = labels.CancelLabelParams
	CreateLabelRequest         = labels.CreateLabelRequest
	CreateLabelJSONRequestBody = labels.CreateLabelJSONRequestBody
	ImageInfo                  = labels.ImageInfo
	LabelAddress               = labels.LabelAddress
	PackageDescription         = labels.PackageDescription
	CustomerReference          = labels.CustomerReference
	LabelResponse              = labels.LabelResponse
	PaymentAuth                = labels.PaymentAuth
)

// Prices API types.
type (
	PricesClient                   = prices.ClientWithResponses
	PricesRequestEditor            = prices.RequestEditorFn
	BaseRatesSearchRequest         = prices.BaseRatesSearchRequest
	BaseRatesSearchResponse        = prices.BaseRatesSearchResponse
	SearchBaseRatesJSONRequestBody = prices.SearchBaseRatesJSONRequestBody
)

// Tracking API types.
type (
	TrackingClient           = tracking.ClientWithResponses
	TrackingRequestEditor    = tracking.RequestEditorFn
	GetPackageTrackingParams = tracking.GetPackageTrackingParams
	TrackingResponse         = tracking.TrackingResponse
	ProofOfDelivery          = tracking.ProofOfDelivery

	GetPackageTrackingParamsExpand = tracking.GetPackageTrackingParamsExpand
)

const (
	TrackingExpandDetail  = tracking.DETAIL
	TrackingExpandSummary = tracking.SUMMARY
)
