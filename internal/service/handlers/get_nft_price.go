package handlers

import (
	"github.com/dl-nft-books/price-svc/internal/service/coingecko/models"
	"github.com/dl-nft-books/price-svc/internal/service/requests"
	"github.com/dl-nft-books/price-svc/resources"
	"github.com/spf13/cast"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"net/http"
)

func GetNftPrice(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewGetPriceRequest(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}
	coingeckoContract := request.Contract
	if mockedToken, ok := MockedNfts(r)[request.Contract]; ok {
		coingeckoContract = mockedToken
	}
	price, err := getNftPrice(r, request.Platform, coingeckoContract)
	if err != nil {
		ape.RenderErr(w, problems.InternalError())
		Log(r).WithError(err).Error("failed to get price")
		return
	}

	if price.Usd == 0 {
		ape.RenderErr(w, problems.NotFound())
		return
	}

	response := resources.NftPriceResponse{
		Data: resources.NftPrice{
			Key: resources.Key{
				ID:   request.Contract,
				Type: resources.NFT_PRICE,
			},
			Attributes: resources.NftPriceAttributes{
				NativeCurrency: price.NativeCurrency,
				Usd:            price.Usd,
			},
		},
	}
	ape.Render(w, response)
}

func getNftPrice(r *http.Request, platform, contract string) (*models.FloorPrice, error) {
	if mockedPlatform, ok := MockedPlatforms(r)[platform]; ok {
		tokenPrice := cast.ToFloat64(mockedPlatform.PricePerOneToken)
		nftPrice := cast.ToFloat64(mockedPlatform.PricePerOneNft)
		if tokenPrice > 0 && nftPrice > 0 {
			return &models.FloorPrice{
				NativeCurrency: float32(nftPrice) / float32(tokenPrice),
				Usd:            float32(nftPrice),
			}, nil
		}
	}
	return Coingecko(r).GetNftPrice(platform, contract)
}
