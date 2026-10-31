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
	StandardizedAddress    = addresses.DomesticAddress
	AddressAdditionalInfo  = addresses.AddressAdditionalInfo
)

// Labels API types (portal Labels v3).
type (
	LabelsClient               = labels.ClientWithResponses
	LabelsRequestEditor        = labels.RequestEditorFn
	CreateLabelParams          = labels.PostLabelParams
	GetLabelParams             = labels.PostLabelReprintByTrackingParams
	CancelLabelParams          = labels.DeleteLabelByTrackingParams
	CreateLabelRequest         = labels.LabelRequest
	CreateLabelJSONRequestBody = labels.PostLabelJSONRequestBody
	ImageInfo                  = labels.ImageInfo
	LabelRequestImageInfo      = labels.LabelRequest_ImageInfo
	LabelAddress               = labels.DomesticLabelAddress
	LabelToAddress             = labels.DomesticLabelToAddress
	PackageDescription         = labels.DomesticPackageDescription
	PackageDescriptionUnion    = labels.LabelRequest_PackageDescription
	CustomerReference          = labels.DomesticCustomerReference
	LabelResponse              = labels.LabelMultiPartResponse
	LabelMetadata              = labels.LabelMetadata
	LabelReprintRequest        = labels.LabelReprintRequest
	LabelReprintImageInfo      = labels.LabelReprintImageInfo
	PaymentAuth                = labels.XPaymentAuthorizationToken
	TrackingNumber             = labels.TrackingNumber
	Email                      = labels.Email
	Phone                      = labels.Phone
	State                      = labels.State

	PackageMailClass                    = labels.DomesticPackageDescriptionMailClass
	PackageRateIndicator                = labels.DomesticPackageDescriptionRateIndicator
	PackageWeightUOM                    = labels.DomesticPackageDescriptionWeightUOM
	PackageDimensionsUOM                = labels.DomesticPackageDescriptionDimensionsUOM
	PackageProcessingCategory           = labels.DomesticPackageDescriptionProcessingCategory
	PackageDestinationEntryFacilityType = labels.DomesticPackageDescriptionDestinationEntryFacilityType
	LabelImageType                      = labels.LabelRequestImageInfoImageType
	LabelLabelType                      = labels.LabelRequestImageInfoLabelType
	ReprintImageType                    = labels.LabelReprintImageInfoImageType
)

// PackageDescriptionFrom builds the OpenAPI union for LabelRequest.packageDescription.
func PackageDescriptionFrom(v PackageDescription) (PackageDescriptionUnion, error) {
	var u PackageDescriptionUnion
	err := u.FromDomesticPackageDescription(v)
	return u, err
}

// Prices API types (portal Domestic Prices).
type (
	PricesClient                               = prices.ClientWithResponses
	PricesRequestEditor                        = prices.RequestEditorFn
	BaseRatesQuery                             = prices.BaseRatesQuery
	BaseRateOption                             = prices.BaseRateOption
	PostBaseRatesSearchJSONRequestBody         = prices.PostBaseRatesSearchJSONRequestBody
	MailClass                                  = prices.MailClass
	PriceType                                  = prices.PriceType
	BaseRatesQueryProcessingCategory           = prices.BaseRatesQueryProcessingCategory
	BaseRatesQueryRateIndicator                = prices.BaseRatesQueryRateIndicator
	BaseRatesQueryDestinationEntryFacilityType = prices.BaseRatesQueryDestinationEntryFacilityType
	BaseRatesQueryAccountType                  = prices.BaseRatesQueryAccountType

	// Backward-compatible aliases for the focused prices seed.
	BaseRatesSearchRequest         = prices.BaseRatesQuery
	BaseRatesSearchResponse        = prices.BaseRateOption
	SearchBaseRatesJSONRequestBody = prices.PostBaseRatesSearchJSONRequestBody
)

// Tracking API types (portal Tracking v3r2 — POST body).
type (
	TrackingClient         = tracking.ClientWithResponses
	TrackingRequestEditor  = tracking.RequestEditorFn
	TrackingRequest        = tracking.TrackingRequest
	TrackingRequestBody    = tracking.TrackingRequestBody
	TrackingResponse       = tracking.TrackingDetails
	TrackingDetail         = tracking.TrackingDetail
	ProofOfDelivery        = tracking.ProofOfDeliveryRequest
	GetPackageTrackingBody = tracking.GetPackageTrackingJSONRequestBody
)
