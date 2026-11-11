package auth

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/meshcloud/meshstack-cli/cmd/internal"
	"github.com/meshcloud/meshstack-cli/cmd/internal/prompt"
	"github.com/meshcloud/meshstack-cli/internal/auth"
	"github.com/meshcloud/meshstack-cli/internal/profile"
	"github.com/meshcloud/meshstack-cli/pkg/setting"
)

func newLogout() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out by removing the stored credentials of a profile",
		Long: `Remove the stored credentials of a profile. For a browser login, this also ends the login's session at meshStack, so its tokens stop working. Your browser stays signed in to meshStack.

Without --` + internal.ProfileFlag.Name.String() + `, the command asks which of the stored profiles for the endpoint to log out of.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			currentProfile, profiles, err := profile.ResolveProfile(ctx,
				append(internal.SettingSources(), newProfileSelectionSource(prompt.New(cmd.InOrStdin(), cmd.ErrOrStderr()), false)))
			if len(profiles.Profiles) == 0 && errors.Is(err, setting.ErrNoSourceProvidedValue) {
				slog.WarnContext(ctx, "Nothing to log out of: no profile is stored")
				return nil
			} else if err != nil {
				return err
			}
			if currentProfile.Name == "" {
				slog.WarnContext(ctx, "Nothing to log out of: no stored profile is used")
				return nil
			}
			if err := auth.Logout(ctx, currentProfile, internal.ResolveClientOptions()); err != nil {
				return err
			}
			slog.InfoContext(ctx, fmt.Sprintf("Logged out of profile '%s'", currentProfile))
			return nil
		},
	}
	return cmd
}
