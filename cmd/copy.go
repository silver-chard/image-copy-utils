/*
Package cmd
Copyright © 2025 xie.zhida@icloud.com
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/silver-chard/image-copy-utils/image"
	"github.com/silver-chard/image-copy-utils/utils"
)

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().StringVarP(&srcImage, "src-image", "s", "", "source image and tag")
	copyCmd.Flags().StringVarP(&dstImage, "dst-image", "d", "", "destination image and tag")
	copyCmd.Flags().StringVar(&srcProxy, "src-proxy", "", "[option] source proxy")
	copyCmd.Flags().StringVar(&dstProxy, "dst-proxy", "", "[option] destination proxy")
	copyCmd.Flags().StringVar(&srcAuth, "src-auth", "", "[option] source auth, support google cloud service account json file or user:password format")
	copyCmd.Flags().StringVar(&dstAuth, "dst-auth", "", "[option] destination auth, support google cloud service account json file or user:password format")
	copyCmd.Flags().IntVarP(&parallelism, "parallelism", "p", 10, "parallelism for copy, default is 10")
	copyCmd.Flags().BoolVar(&debug, "debug", false, "print i/o and speed info when copy")
}

var srcImage, dstImage string
var srcProxy, dstProxy string
var srcAuth, dstAuth string
var debug bool
var parallelism int

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use: "copy",
	Short: "copy --src-image xxx.xxx:xxx --dst-image yyy.xxx:xxx --src-proxy proxy-1.path --dst-proxy proxy-2.path --src-auth auth." +
		"json|user@password --dst-auth auth.json|user@password",
	Long: `copy image from src-image to dst-image with proxy`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(srcImage)*len(dstImage) == 0 {
			return errors.New("no image provided")
		}
		srcAuthenticator, err := image.GenAuthenticator(srcAuth)
		if err != nil {
			return fmt.Errorf("generate src auth failed: %w", err)
		}
		dstAuthenticator, err := image.GenAuthenticator(dstAuth)
		if err != nil {
			return fmt.Errorf("generate dst auth failed: %w", err)
		}
		srcStats, srcDialer, err := image.GenHTTPStatRounderTripper(srcProxy)
		if err != nil {
			return fmt.Errorf("generate src dialer failed: %w", err)
		}
		dstStats, dstDialer, err := image.GenHTTPStatRounderTripper(dstProxy)
		if err != nil {
			return fmt.Errorf("generate dst dialer failed: %w", err)
		}
		go debugFunc(srcStats, dstStats)
		return image.CopyImage(context.Background(), srcImage, dstImage,
			image.SetSrcImageAuth(srcAuthenticator), image.SetDstImageAuth(dstAuthenticator),
			image.SetSrcRounderTripper(srcDialer), image.SetDstRounderTripper(dstDialer),
			image.SetSrcParallelism(parallelism), image.SetDstParallelism(parallelism),
		)
	},
}

func debugFunc(srcDialer, dstDialer *image.Statistic) {
	if !debug {
		return
	}
	t := time.NewTicker(time.Second)
	for {
		<-t.C
		fmt.Print("\033[2J\033[H")
		log("src", srcDialer)
		log("dst", dstDialer)
	}
}

func log(prefix string, stats *image.Statistic) {
	downBytes, upBytes, dial, retry, startTime := stats.GetStatistic()
	if uint64(time.Since(startTime).Seconds()) == 0 {
		return
	}
	fmt.Printf(strings.Repeat("-", 20))
	fmt.Printf("[%s]\n", prefix)
	fmt.Printf("down: %-10s up: %-10s downSpeed: %10s/s upSpeed: %10s/s\n",
		utils.HumanSize(downBytes), utils.HumanSize(upBytes),
		utils.HumanSize(downBytes/uint64(time.Since(startTime).Seconds())),
		utils.HumanSize(upBytes/uint64(time.Since(startTime).Seconds())),
	)
	fmt.Printf("dial: %d retry: %d\n", dial, retry)
}
