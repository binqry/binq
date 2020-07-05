package cli

type confirmRunner interface {
	runner
	getConfirmOpts() confirmFlavor
}

type confirmFlavor interface {
	getYes() *bool
}

type confirmCmd struct {
	*commonCmd
	option *confirmOpts
}

type confirmOpts struct {
	yes *bool
	*commonOpts
}

func (cmd *confirmCmd) getConfirmOpts() confirmFlavor {
	return cmd.option
}

func (opt *confirmOpts) getYes() (y *bool) {
	return opt.yes
}
