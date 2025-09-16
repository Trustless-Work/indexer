package escrow

type Roles struct {
	Approver        string `json:"approver"`
	ServiceProvider string `json:"serviceProvider"`
	PlatformAddress string `json:"platformAddress"`
	ReleaseSigner   string `json:"releaseSigner"`
	DisputeResolver string `json:"disputeResolver"`
	Receiver        string `json:"receiver"`
}

type Trustline struct {
	Address  string `json:"address"`
	Decimals int64  `json:"decimals"`
}

type SingleReleaseJSON struct {
	ContractID   string `json:"contractId"` // requerido (PK)
	Signer       string `json:"signer"`
	EngagementID string `json:"engagementId"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Roles        Roles  `json:"roles"`
	Amount       int64  `json:"amount"`
	PlatformFee  int64  `json:"platformFee"`
	Milestones   []struct {
		Description string `json:"description"`
	} `json:"milestones"`
	Trustline    Trustline `json:"trustline"`
	ReceiverMemo int64     `json:"receiverMemo"`
}

type MultiReleaseJSON struct {
	ContractID   string `json:"contractId"` // requerido (PK)
	Signer       string `json:"signer"`
	EngagementID string `json:"engagementId"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Roles        Roles  `json:"roles"`
	PlatformFee  int64  `json:"platformFee"`
	Milestones   []struct {
		Description string `json:"description"`
		Amount      int64  `json:"amount"`
	} `json:"milestones"`
	Trustline    Trustline `json:"trustline"`
	ReceiverMemo int64     `json:"receiverMemo"`
}
