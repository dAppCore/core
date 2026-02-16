package crypt

import "forge.lthn.ai/core/cli/pkg/cli"

func init() {
	cli.RegisterCommands(AddCryptCommands)
}

// AddCryptCommands registers the 'crypt' command group and all subcommands.
func AddCryptCommands(root *cli.Command) {
	cryptCmd := &cli.Command{
		Use:   "crypt",
		Short: "Cryptographic utilities",
		Long:  "Encrypt, decrypt, hash, and checksum files and data.",
	}
	root.AddCommand(cryptCmd)

	addHashCommand(cryptCmd)
	addEncryptCommand(cryptCmd)
	addKeygenCommand(cryptCmd)
	addChecksumCommand(cryptCmd)
}
