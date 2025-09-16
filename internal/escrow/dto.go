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
	Name     string `json:"name"`
}

type SingleReleaseJSON struct {
	ContractID       string `json:"contractId"` // requerido (PK)
	ContractBaseID   string `json:"contractBaseId"`
	Signer           string `json:"signer"`
	EngagementID     string `json:"engagementId"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	Roles            Roles  `json:"roles"`
	AmountRaw        int64  `json:"amountRaw"`    // NUMERIC(39,0) as int64 
	BalanceRaw       int64  `json:"balanceRaw"`   // NUMERIC(39,0) as int64
	Amount           int64  `json:"amount"`       // For backward compatibility
	Balance          int64  `json:"balance"`      // For backward compatibility  
	PlatformFee      int64  `json:"platformFee"`
	Milestones       []struct {
		Description string `json:"description"`
	} `json:"milestones"`
	Trustline        Trustline `json:"trustline"`
	ReceiverMemo     int64     `json:"receiverMemo"`
}

type MultiReleaseJSON struct {
	ContractID     string `json:"contractId"` // requerido (PK)
	ContractBaseID string `json:"contractBaseId"`
	Signer         string `json:"signer"`
	EngagementID   string `json:"engagementId"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Roles          Roles  `json:"roles"`
	PlatformFee    int64  `json:"platformFee"`
	Milestones     []struct {
		Description string `json:"description"`
		Amount      int64  `json:"amount"`
	} `json:"milestones"`
	Trustline      Trustline `json:"trustline"`
	ReceiverMemo   int64     `json:"receiverMemo"`
}
