package libfuse

import (
	"encoding/json"
	"time"

	"bazil.org/fuse"

	"github.com/keybase/kbfs/libkbfs"
	"golang.org/x/net/context"
)

// StatusFileName is the name of the KBFS status file -- it can be
// reached anywhere within a top-level folder.
const StatusFileName = ".kbfs_status"

type kbfsStatus struct {
	CurrentUser string
	IsConnected bool
	UsageBytes  int64
	LimitBytes  int64
}

func getEncodedFolderStatus(ctx context.Context, fs *FS,
	folderBranch *libkbfs.FolderBranch) (
	data []byte, t time.Time, err error) {

	var status libkbfs.FolderBranchStatus
	status, _, err = fs.config.KBFSOps().
		Status(ctx, *folderBranch)
	if err != nil {
		return nil, time.Time{}, err
	}

	data, err = json.MarshalIndent(status, "", "  ")
	if err != nil {
		return nil, time.Time{}, err
	}

	data = append(data, '\n')
	return data, time.Time{}, err
}

func getEncodedStatus(ctx context.Context, fs *FS) (
	data []byte, t time.Time, err error) {
	username, _, _ := fs.config.KBPKI().GetCurrentUserInfo(ctx)
	quotaInfo, err := fs.config.BlockServer().GetUserQuotaInfo(ctx)
	var usageBytes int64
	var limitBytes int64
	if err != nil {
		usageBytes = quotaInfo.Total.UsageBytes
		limitBytes = quotaInfo.Limit
	}
	data, err = json.MarshalIndent(kbfsStatus{
		CurrentUser: username.String(),
		IsConnected: fs.config.MDServer().IsConnected(),
		UsageBytes:  usageBytes,
		LimitBytes:  limitBytes,
	}, "", "  ")
	if err != nil {
		return nil, t, err
	}
	data = append(data, '\n')
	return data, t, err
}

// NewStatusFile returns a special read file that contains a text
// representation of the status of the current TLF.
func NewStatusFile(fs *FS, folderBranch *libkbfs.FolderBranch, resp *fuse.LookupResponse) *SpecialReadFile {
	resp.EntryValid = 0
	return &SpecialReadFile{
		read: func(ctx context.Context) ([]byte, time.Time, error) {
			if folderBranch == nil {
				return getEncodedStatus(ctx, fs)
			}
			return getEncodedFolderStatus(ctx, fs, folderBranch)
		},
	}
}
