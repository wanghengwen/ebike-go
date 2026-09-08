package auth

import "ebike-auth-go/internal/pkg/utils"

// Shared test constants aligned with Java ebike-auth-client TestJwt and providers.

const (
	testTenantID   = "1"
	testTraceID    = "trace-integration-test"
	testDeviceID   = "device-test-001"
	testPlatform   = "ios"
	testJwtSecret  = "dbgpfbktyq0p2fowbll4m22xjom6bhc8g0pwmb9b0pj7g67p414ctndum3e0ectf"
	testClientPass = "xyy"

	// Java TestJwt.test()
	javaAppleIdentityToken = "eyJraWQiOiJlWGF1bm1MIiwiYWxnIjoiUlMyNTYifQ.eyJpc3MiOiJodHRwczovL2FwcGxlaWQuYXBwbGUuY29tIiwiYXVkIjoiY29tLmRpeWl5aW4ub25saW5lNTMiLCJleHAiOjE1OTc2NTAxNzQsImlhdCI6MTU5NzY0OTU3NCwic3ViIjoiMDAxMzc3LmQ0ZDVmMTAwODQ0ZTQzZjdiMWM1OWRiMzUyZWZkZmI4LjAyNTkiLCJjX2hhc2giOiJkTDVRdld2VTNjVHBxczNSazlUTnRBIiwiZW1haWwiOiI0OTk4OTY1MDdAcXEuY29tIiwiZW1haWxfdmVyaWZpZWQiOiJ0cnVlIiwiYXV0aF90aW1lIjoxNTk3NjQ5NTc0LCJub25jZV9zdXBwb3J0ZWQiOnRydWV9.hM9HjNsMJW2PjYP7SfbzF-GqOt0VnMjYGq4BoU68rkQ-K2lPp_ae5ziX6Bbr3WHg6cc3Z8OzGO63OfExvSj9gQTR596CZLvNGXhbI3piTK6597-cYsPCTbY7xHxgdHLuL8XhD-9dXPn9rouVYu4QA18JBQG1Q4sGsRzLEJ5DjOM9x1bkBz4Vu_5LEOefHFHkWN_RPCh_AOJGviDzm81kTkCTWn8jpm0tGdevMR93MOf44f7bjP2T8yezl0Vbv09TrnkdAqG0BsihCD0VN9JV7X2eagyumoxTdFfoRiOflFKAaQqohVzcqy9tHOGm_6w5h8bsRCmtBC4PnqIFqNy_AQ"
	javaApplePublicKeyN    = "4dGQ7bQK8LgILOdLsYzfZjkEAoQeVC_aqyc8GC6RX7dq_KvRAQAWPvkam8VQv4GK5T4ogklEKEvj5ISBamdDNq1n52TpxQwI2EqxSk7I9fKPKhRt4F8-2yETlYvye-2s6NeWJim0KBtOVrk0gWvEDgd6WOqJl_yt5WBISvILNyVg1qAAM8JeX6dRPosahRVDjA52G2X-Tip84wqwyRpUlq2ybzcLh3zyhCitBOebiRWDQfG26EH9lTlJhll-p_Dg8vAXxJLIJ4SNLcqgFeZe4OfHLgdzMvxXZJnPp_VgmkcpUdRotazKZumj6dBPcXI_XID4Z4Z3OM1KrZPJNdUhxw"
	javaApplePublicKeyE    = "AQAB"

	// Java PhoneSecretAuthenticationProvider.main()
	javaPhoneSecretPhone = "+86-13381458187"

	// Java TestJwt.test1() UnionPay fixtures (for reference / future live tests)
	javaUnionPayTenantID = "250"
	javaUnionPayAppID    = "38c6849a093b4545917029a2ee823dd5"
	javaUnionPaySecret   = "fb6b3b3247aa47e1bc6be201dace9cbd"
	javaUnionPayDcSecret = "25326b85bc45738a19070d86d5a4f85825326b85bc45738a"

	javaGooglePublicKeyKid = "d4e06ceb22b01be56c213c98540ab563bff5a58c"
)

func javaClientPhoneSecret(tenantID string) string {
	return utils.Sha256(javaPhoneSecretPhone + "_xyy@2022@" + tenantID)
}

func javaBusinessPhoneSecret(phone string) string {
	return utils.Sha256(phone + "_xyy@2022")
}
