package cos_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sut "github.com/rancher-sandbox/ele-testhelpers/vm"
)

var _ = Describe("cOS Installer tests", func() {
	var s *sut.SUT

	BeforeEach(func() {
		s = sut.NewSUT()
		s.EventuallyConnects()
	})

	Context("Using bios", func() {
		BeforeEach(func() {

			s.EmptyDisk("/dev/sda")
			// Only reboot if we boot from other than the CD to speed up test preparation
			if s.BootFrom() != sut.LiveCD {
				By("Reboot to make sure we boot from CD")
				s.Reboot()
			}

			// Assert we are booting from CD before running the tests
			By("Making sure we booted from CD")
			ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.LiveCD))
			out, err := s.Command("grep /dev/sr /etc/mtab")
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(ContainSubstring("iso9660"))
			out, err = s.Command("df -h /")
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(ContainSubstring("LiveOS_rootfs"))

			// squashfs comes from a command line flag at suite level
			if squashfs {
				By("Setting the squashfs recovery install")
				err = s.SendFile("../assets/squashed_recovery.yaml", "/etc/elemental/config.d/install_recovery.yaml", "0770")
				Expect(err).ToNot(HaveOccurred())
			}
		})

		AfterEach(func() {
			if CurrentSpecReport().Failed() {
				s.GatherAllLogs()
			}
		})

		Context("install source tests", func() {
			It("from iso", func() {
				By("Running the elemental install")
				out, err := s.Command(s.ElementalCmd("install", "/dev/sda"))
				Expect(err).To(BeNil())
				Expect(out).To(ContainSubstring("Installing GRUB.."))
				Expect(out).To(ContainSubstring("Mounting disk partitions"))
				Expect(out).To(ContainSubstring("Partitioning device..."))
				Expect(out).To(ContainSubstring("Unmounting disk partitions"))
				Expect(out).To(ContainSubstring("Running after-install hook"))

				if squashfs {
					// Check the squashfs image is used as recovery
					Expect(out).To(ContainSubstring("/run/initramfs/live/rootfs.squashfs into /run/cos/recovery/cOS/recovery.squashfs"))
				}
				// Reboot so we boot into the just installed cos
				s.Reboot()
				By("Checking we booted from the installed cOS")
				ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Active))
			})
			PIt("from url", func() {})
			It("from docker image", func() {
				By("Running the elemental install")
				out, err := s.Command(s.ElementalCmd("install", "--system.uri", fmt.Sprintf("docker:%s:cos-system-%s", s.GetArtifactsRepo(), s.TestVersion), "/dev/sda"))
				Expect(err).To(BeNil())
				Expect(out).To(ContainSubstring("Installing GRUB.."))
				Expect(out).To(ContainSubstring("Mounting disk partitions"))
				Expect(out).To(ContainSubstring("Partitioning device..."))
				Expect(out).To(ContainSubstring("Unmounting disk partitions"))
				Expect(out).To(ContainSubstring("Running after-install hook"))
				if squashfs {
					// Check the squashfs image is used as recovery
					Expect(out).To(ContainSubstring("/run/initramfs/live/rootfs.squashfs into /run/cos/recovery/cOS/recovery.squashfs"))
				}
				// Reboot so we boot into the just installed cos
				s.Reboot()
				By("Checking we booted from the installed cOS")
				ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Active))
				Expect(s.GetOSRelease("VERSION")).To(Equal(s.TestVersion))
			})
		})

		Context("partition layout tests", func() {
			Context("with partition layout", func() {
				It("Forcing GPT", func() {
					err := s.SendFile("../assets/custom_partitions.yaml", "/etc/elemental/config.d/custom_partitions.yaml", "0770")
					By("Running the elemental installer with a layout file")
					Expect(err).To(BeNil())
					out, err := s.Command(s.ElementalCmd("install", "--force-gpt", "/dev/sda"))
					Expect(err).To(BeNil())
					Expect(out).To(ContainSubstring("Installing GRUB.."))
					Expect(out).To(ContainSubstring("Mounting disk partitions"))
					Expect(out).To(ContainSubstring("Partitioning device..."))
					Expect(out).To(ContainSubstring("Unmounting disk partitions"))
					Expect(out).To(ContainSubstring("Running after-install hook"))
					if squashfs {
						// Check the squashfs image is used as recovery
						Expect(out).To(ContainSubstring("/run/initramfs/live/rootfs.squashfs into /run/cos/recovery/cOS/recovery.squashfs"))
					}
					s.Reboot()
					By("Checking we booted from the installed cOS")
					ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Active))

					// check partition values
					// Values have to match the yaml under ../assets/layout.yaml
					// That is the file that the installer uses so partitions should match those values
					disk := s.GetDiskLayout("/dev/sda")

					for _, part := range []sut.PartitionEntry{
						{
							Label:  "COS_STATE",
							Size:   8192,
							FsType: sut.Ext4,
						},
						{
							Label:  "COS_OEM",
							Size:   10,
							FsType: sut.Ext4,
						},
						{
							Label:  "COS_RECOVERY",
							Size:   4000,
							FsType: sut.Ext2,
						},
						{
							Label:  "COS_PERSISTENT",
							Size:   100,
							FsType: sut.Ext2,
						},
					} {
						CheckPartitionValues(disk, part)
					}
				})

				It("No GPT", func() {
					err := s.SendFile("../assets/custom_partitions.yaml", "/etc/elemental/config.d/custom_partitions.yaml", "0770")
					By("Running the elemental install with a layout file")
					Expect(err).To(BeNil())
					out, err := s.Command(s.ElementalCmd("install", "/dev/sda"))
					Expect(err).To(BeNil())
					Expect(out).To(ContainSubstring("Installing GRUB.."))
					Expect(out).To(ContainSubstring("Mounting disk partitions"))
					Expect(out).To(ContainSubstring("Partitioning device..."))
					Expect(out).To(ContainSubstring("Unmounting disk partitions"))
					Expect(out).To(ContainSubstring("Running after-install hook"))
					if squashfs {
						// Check the squashfs image is used as recovery
						Expect(out).To(ContainSubstring("/run/initramfs/live/rootfs.squashfs into /run/cos/recovery/cOS/recovery.squashfs"))
					}
					s.Reboot()
					By("Checking we booted from the installed cOS")
					ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Active))

					// check partition values
					// Values have to match the yaml under ../assets/layout.yaml
					// That is the file that the installer uses so partitions should match those values
					disk := s.GetDiskLayout("/dev/sda")

					for _, part := range []sut.PartitionEntry{
						{
							Label:  "COS_STATE",
							Size:   8192,
							FsType: sut.Ext4,
						},
						{
							Label:  "COS_OEM",
							Size:   10,
							FsType: sut.Ext4,
						},
						{
							Label:  "COS_RECOVERY",
							Size:   4000,
							FsType: sut.Ext2,
						},
						{
							Label:  "COS_PERSISTENT",
							Size:   100,
							FsType: sut.Ext2,
						},
					} {
						CheckPartitionValues(disk, part)
					}
				})
			})
		})

		Context("efi/gpt tests", func() {
			It("forces gpt", func() {
				By("Running the installer")
				out, err := s.Command(s.ElementalCmd("install", "--force-gpt", "/dev/sda"))
				Expect(err).To(BeNil())
				Expect(out).To(ContainSubstring("Installing GRUB.."))
				Expect(out).To(ContainSubstring("Mounting disk partitions"))
				Expect(out).To(ContainSubstring("Partitioning device..."))
				Expect(out).To(ContainSubstring("Unmounting disk partitions"))
				Expect(out).To(ContainSubstring("Running after-install hook"))
				if squashfs {
					// Check the squashfs image is used as recovery
					Expect(out).To(ContainSubstring("/run/initramfs/live/rootfs.squashfs into /run/cos/recovery/cOS/recovery.squashfs"))
				}
				s.Reboot()
				By("Checking we booted from the installed cOS")
				ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Active))
			})

			It("forces efi", func() {
				By("Running the installer")
				out, err := s.Command(s.ElementalCmd("install", "--force-efi", "/dev/sda"))
				Expect(err).To(BeNil())
				Expect(out).To(ContainSubstring("Installing GRUB.."))
				Expect(out).To(ContainSubstring("Mounting disk partitions"))
				Expect(out).To(ContainSubstring("Partitioning device..."))
				Expect(out).To(ContainSubstring("Unmounting disk partitions"))
				Expect(out).To(ContainSubstring("Running after-install hook"))
				if squashfs {
					// Check the squashfs image is used as recovery
					Expect(out).To(ContainSubstring("/run/initramfs/live/rootfs.squashfs into /run/cos/recovery/cOS/recovery.squashfs"))
				}
				s.Reboot()
				// We are on a bios system, we should not be able to boot from an EFI installed system!
				By("Checking we booted from the CD")
				ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.LiveCD))
			})
		})
		Context("config file tests", func() {
			It("uses a proper config file", func() {
				err := s.SendFile("../assets/hostname.yaml", "/tmp/config.yaml", "0770")
				By("Running the elemental install with a config file")
				Expect(err).To(BeNil())
				By("Running the installer")
				out, err := s.Command(s.ElementalCmd("install", "--cloud-init", "/tmp/config.yaml", "/dev/sda"))
				Expect(err).To(BeNil())
				Expect(out).To(ContainSubstring("Installing GRUB.."))
				Expect(out).To(ContainSubstring("Mounting disk partitions"))
				Expect(out).To(ContainSubstring("Partitioning device..."))
				Expect(out).To(ContainSubstring("Unmounting disk partitions"))
				Expect(out).To(ContainSubstring("Running after-install hook"))
				if squashfs {
					// Check the squashfs image is used as recovery
					Expect(out).To(ContainSubstring("/run/initramfs/live/rootfs.squashfs into /run/cos/recovery/cOS/recovery.squashfs"))
				}
				s.Reboot()
				By("Checking we booted from the installed cOS")
				ExpectWithOffset(1, s.BootFrom()).To(Equal(sut.Active))
				By("Checking config file was run")
				out, err = s.Command("stat /oem/90_custom.yaml")
				Expect(err).To(BeNil())
				out, err = s.Command("hostname")
				Expect(err).To(BeNil())
				Expect(out).To(ContainSubstring("testhostname"))
			})
		})
	})
})
