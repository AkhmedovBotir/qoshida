package eskiz

import "strings"

// Eskizda tasdiqlangan matnlar. ${code} jo‘natishda haqiqiy kodga almashtiriladi.
const (
	TplMarketplaceRegister = `${code} - Marketplace ilovasidan ro'yxatdan o'tish uchun tasdiqlash kodi. Kod 5 daqiqa amal qiladi. Talab va Taklif Agency.`
	TplMarketplaceLogin    = `${code} - Marketplace ilovasiga kirish uchun tasdiqlash kodi. Kod 5 daqiqa amal qiladi. Talab va Taklif Agency.`
	TplMarketplaceReset    = `${code} - Marketplace ilovasida parol tiklash uchun tasdiqlash kodi. Kod 5 daqiqa amal qiladi. Talab va Taklif Agency.`
	TplDeviceVerify        = `${code} - Yangi qurilmani tasdiqlash uchun tasdiqlash kodi. Kod 5 daqiqa amal qiladi. Talab va Taklif Agency.`
	TplContragentPassword  = `${code} - Kontragent hisobi uchun parol o'rnatish kodi. Kod 5 daqiqa amal qiladi. Talab va Taklif Agency.`
	TplPunktPassword       = `${code} - Punkt hisobi uchun parol o'rnatish kodi. Kod 5 daqiqa amal qiladi. Talab va Taklif Agency.`
	TplAgentPassword       = `${code} - Agent hisobi uchun parol o'rnatish kodi. Kod 5 daqiqa amal qiladi. Talab va Taklif Agency.`
	TplDeliveryPassword    = `${code} - Yetkazib beruvchi hisobi uchun parol o'rnatish kodi. Kod 5 daqiqa amal qiladi. Talab va Taklif Agency.`
	TplManagerPassword     = `${code} - Menejer hisobi uchun parol o'rnatish kodi. Kod 5 daqiqa amal qiladi. Talab va Taklif Agency.`
	TplLocalShopPassword     = `${code} - Maxalla do'koni hisobi uchun parol o'rnatish kodi. Kod 5 daqiqa amal qiladi. Talab va Taklif Agency.`
	TplShopDirectorPassword  = `${code} - Savdo uyi rahbari hisobi uchun parol o'rnatish kodi. Kod 5 daqiqa amal qiladi. Talab va Taklif Agency.`
	TplServicePassword       = `${code} - Xizmat ko'rsatuvchi hisobi uchun parol o'rnatish kodi. Kod 5 daqiqa amal qiladi. Talab va Taklif Agency.`
	TplSellerPassword        = `${code} - Sotuvchi hisobi uchun parol o'rnatish kodi. Kod 5 daqiqa amal qiladi. Talab va Taklif Agency.`
)

type Template string

const (
	TemplateMarketplaceRegister Template = "marketplace_register"
	TemplateMarketplaceLogin    Template = "marketplace_login"
	TemplateMarketplaceReset    Template = "marketplace_reset"
	TemplateDeviceVerify        Template = "device_verify"
	TemplateContragentPassword  Template = "contragent_password"
	TemplatePunktPassword       Template = "punkt_password"
	TemplateAgentPassword       Template = "agent_password"
	TemplateDeliveryPassword    Template = "delivery_password"
	TemplateManagerPassword     Template = "manager_password"
	TemplateLocalShopPassword    Template = "local_shop_password"
	TemplateShopDirectorPassword Template = "shop_director_password"
	TemplateServicePassword      Template = "service_password"
	TemplateSellerPassword       Template = "seller_password"
)

var templates = map[Template]string{
	TemplateMarketplaceRegister: TplMarketplaceRegister,
	TemplateMarketplaceLogin:    TplMarketplaceLogin,
	TemplateMarketplaceReset:    TplMarketplaceReset,
	TemplateDeviceVerify:        TplDeviceVerify,
	TemplateContragentPassword:  TplContragentPassword,
	TemplatePunktPassword:       TplPunktPassword,
	TemplateAgentPassword:       TplAgentPassword,
	TemplateDeliveryPassword:    TplDeliveryPassword,
	TemplateManagerPassword:     TplManagerPassword,
	TemplateLocalShopPassword:    TplLocalShopPassword,
	TemplateShopDirectorPassword: TplShopDirectorPassword,
	TemplateServicePassword:      TplServicePassword,
	TemplateSellerPassword:       TplSellerPassword,
}

func Render(tpl Template, code string) (string, error) {
	text, ok := templates[tpl]
	if !ok {
		return "", ErrUnknownTemplate
	}
	return strings.ReplaceAll(text, "${code}", code), nil
}
