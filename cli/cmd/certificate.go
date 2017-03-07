package cmd

import (
	"fmt"

	"github.com/jsimonetti/tlstun/cert"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(certificateCmd)
	certificateCmd.AddCommand(caCertCmd)
	certificateCmd.AddCommand(serverCertCmd)
	certificateCmd.AddCommand(clientCertCmd)

	certificateCmd.PersistentFlags().String("certfile", "", "Filename to output certificate to")
	certificateCmd.PersistentFlags().String("keyfile", "", "Filename to output key to")
	certificateCmd.PersistentFlags().String("name", "", "Name of the certificate")
	certificateCmd.PersistentFlags().String("cacert", "", "Name of the CA certificate")
	certificateCmd.PersistentFlags().String("cakey", "", "Name of the CA key")

}

var certificateCmd = &cobra.Command{
	Use:   "certificate",
	Short: "Generate certificates",
	Long: `Generate certificates for a TLSTun certificate authority.

Used to build a ca, client or server certificate using (current) safe ciphers.`,
}

var caCertCmd = &cobra.Command{
	Use:   "ca",
	Short: "Generate CA certificate",
	Run:   caCertGenerate,
}

func caCertGenerate(cmd *cobra.Command, args []string) {
	fmt.Printf("Generating CA certificate ...")

	certf := certificateCmd.PersistentFlags().Lookup("certfile").Value.String()
	keyf := certificateCmd.PersistentFlags().Lookup("keyfile").Value.String()

	cacert, err := cert.CreateCaCertificate()
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	err = cacert.CertToFile(certf)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	err = cacert.KeyToFile(keyf)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	fmt.Printf(" done.\nCA certificate created at %s\n", certf)
}

var serverCertCmd = &cobra.Command{
	Use:   "server",
	Short: "Generate server certificate signed by CA",
	Run:   serverCertGenerate,
}

func serverCertGenerate(cmd *cobra.Command, args []string) {
	fmt.Printf("Generating Server certificate ...")

	certf := certificateCmd.PersistentFlags().Lookup("certfile").Value.String()
	keyf := certificateCmd.PersistentFlags().Lookup("keyfile").Value.String()
	cafile := certificateCmd.PersistentFlags().Lookup("cacert").Value.String()
	cakey := certificateCmd.PersistentFlags().Lookup("cakey").Value.String()
	name := certificateCmd.PersistentFlags().Lookup("name").Value.String()

	var ca cert.Certificate
	err := ca.FromFile(cafile, cakey)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	servercert, err := cert.CreateServerCertificate(ca, name)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	err = servercert.CertToFile(certf)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	err = servercert.KeyToFile(keyf)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	fmt.Printf("done.\nServer certificate created at %s\n", certf)
}

var clientCertCmd = &cobra.Command{
	Use:   "client",
	Short: "Generate client certificate signed by CA",
	Run:   clientCertGenerate,
}

func clientCertGenerate(cmd *cobra.Command, args []string) {
	fmt.Printf("Generating Client certificate ...")

	certf := certificateCmd.PersistentFlags().Lookup("certfile").Value.String()
	keyf := certificateCmd.PersistentFlags().Lookup("keyfile").Value.String()
	cafile := certificateCmd.PersistentFlags().Lookup("cacert").Value.String()
	cakey := certificateCmd.PersistentFlags().Lookup("cakey").Value.String()
	name := certificateCmd.PersistentFlags().Lookup("name").Value.String()

	var ca cert.Certificate
	err := ca.FromFile(cafile, cakey)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	servercert, err := cert.CreateServerCertificate(ca, name)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	err = servercert.CertToFile(certf)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	err = servercert.KeyToFile(keyf)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	fmt.Printf("done.\nClient certificate created at %s\n", certf)
}
