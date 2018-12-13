package integration_test

import (
	"os/exec"

	"github.com/concourse/atc"
	"github.com/concourse/fly/ui"
	"github.com/fatih/color"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = FDescribe("Fly CLI", func() {
	Describe("find", func() {
		var (
			flyCmd *exec.Cmd
		)

		BeforeEach(func() {
			flyCmd = exec.Command(flyPath, "-t", targetName, "find", "--build-id", "122")
		})

		Context("when containers are returned from the API", func() {
			BeforeEach(func() {
				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/teams/main/containers"),
						ghttp.RespondWithJSONEncoded(200, []atc.Container{
							{
								ID:           "handle-1",
								WorkerName:   "worker-name-1",
								PipelineName: "pipeline-name",
								Type:         "check",
								ResourceName: "git-repo",
							},
							{
								ID:           "early-handle",
								WorkerName:   "worker-name-1",
								PipelineName: "pipeline-name",
								JobName:      "job-name-1",
								BuildName:    "3",
								BuildID:      123,
								Type:         "get",
								StepName:     "git-repo",
								Attempt:      "1.5",
							},
							{
								ID:           "other-handle",
								WorkerName:   "worker-name-2",
								PipelineName: "pipeline-name",
								JobName:      "job-name-2",
								BuildName:    "2",
								BuildID:      122,
								Type:         "task",
								StepName:     "unit-tests",
							},
							{
								ID:         "post-handle",
								WorkerName: "worker-name-3",
								BuildID:    142,
								Type:       "task",
								StepName:   "one-off",
							},
						}),
					),
				)
			})

			It("lists workers associated with the given build id", func() {
				sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(sess).Should(gexec.Exit(0))
				Expect(sess.Out).To(PrintTable(ui.Table{
					Headers: ui.TableRow{
						{Contents: "build id", Color: color.New(color.Bold)},
						{Contents: "worker", Color: color.New(color.Bold)},
					},
					Data: []ui.TableRow{
						{{Contents: "122"}, {Contents: "worker-name-2"}},
					},
				}))
			})

			Context("when the --build-id is not found", func() {
				BeforeEach(func() {
					flyCmd = exec.Command(flyPath, "-t", targetName, "find", "--build-id", "456")
				})

				It("prints that none were found", func() {
					sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(sess).Should(gexec.Exit(0))
					Expect(sess.Out).To(PrintTable(ui.Table{
						Headers: ui.TableRow{
							{Contents: "build id", Color: color.New(color.Bold)},
							{Contents: "worker", Color: color.New(color.Bold)},
						},
						Data: []ui.TableRow{
							{{Contents: "456"}, {Contents: "no worker found"}},
						},
					}))
				})
			})

			Context("when --json is given", func() {
				BeforeEach(func() {
					flyCmd.Args = append(flyCmd.Args, "--json")
				})

				It("prints response in json as stdout", func() {
					sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(sess).Should(gexec.Exit(0))
					Expect(sess.Out.Contents()).To(MatchJSON(`
              {
                "build_id": "122",
                "worker_name": "worker-name-2"
              }
            `))
				})
			})
		})

		Context("and the api returns an internal server error", func() {
			BeforeEach(func() {
				atcServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/teams/main/containers"),
						ghttp.RespondWith(500, ""),
					),
				)
			})

			It("writes an error message to stderr", func() {
				sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(sess).Should(gexec.Exit(1))
				Eventually(sess.Err).Should(gbytes.Say("Unexpected Response"))
			})
		})

		Context("when no --build-id is given", func() {
			BeforeEach(func() {
				flyCmd = exec.Command(flyPath, "-t", targetName, "find")
			})

			It("instructs the user to specify the required flag", func() {
				sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess.Err).Should(gbytes.Say("Must specify --build-id"))
				Eventually(sess).Should(gexec.Exit(1))
			})
		})
	})
})
