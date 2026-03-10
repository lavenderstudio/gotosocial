// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package common

import (
	"context"
	"errors"
	"fmt"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/media"
)

// StoreLocalMedia is a wrapper around CreateMedia() and
// ProcessingMedia{}.Load() with appropriate error responses.
func (p *Processor) StoreLocalMedia(
	ctx context.Context,
	accountID string,
	data media.DataFunc,
	info media.AdditionalMediaInfo,
) (
	*gtsmodel.MediaAttachment,
	gtserror.WithCode,
) {
	// Create a new processing media attachment.
	processing, err := p.media.CreateMedia(ctx,
		accountID,
		data,
		info,
	)
	if err != nil {
		err := gtserror.Newf("error creating media: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Immediately trigger write to storage.
	attachment, err := processing.Load(ctx)
	switch {
	case gtserror.LimitReached(err):
		limit := config.GetMediaLocalMaxSize()
		text := fmt.Sprintf("local media size limit reached: %s", limit)
		return nil, gtserror.NewErrorUnprocessableEntity(err, text)

	case err != nil:
		const text = "error processing media"
		err := gtserror.Newf("error processing media: %w", err)
		return nil, gtserror.NewErrorUnprocessableEntity(err, text)

	case attachment.Type == gtsmodel.FileTypeUnknown:
		text := fmt.Sprintf("could not process %s type media", attachment.File.ContentType)
		return nil, gtserror.NewErrorUnprocessableEntity(errors.New(text), text)
	}

	return attachment, nil
}

// StoreLocalMedia is a wrapper around CreateMedia() and
// ProcessingMedia{}.Load() with appropriate error responses.
func (p *Processor) StoreLocalEmoji(
	ctx context.Context,
	shortcode string,
	data media.DataFunc,
	info media.AdditionalEmojiInfo,
) (
	*gtsmodel.Emoji,
	gtserror.WithCode,
) {
	// Create a new processing emoji media.
	processing, err := p.media.CreateEmoji(ctx,
		shortcode,
		"", // domain = "" -> local
		data,
		info,
	)
	if err != nil {
		err := gtserror.Newf("error creating emoji: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Immediately trigger write to storage.
	emoji, err := processing.Load(ctx)
	switch {
	case gtserror.LimitReached(err):
		limit := config.GetMediaEmojiLocalMaxSize()
		text := fmt.Sprintf("local emoji size limit reached: %s", limit)
		return nil, gtserror.NewErrorUnprocessableEntity(err, text)

	case err != nil:
		const text = "error processing emoji"
		err := gtserror.Newf("error processing emoji %s: %w", shortcode, err)
		return nil, gtserror.NewErrorUnprocessableEntity(err, text)
	}

	return emoji, nil
}

// LoadAccountMedia ensures that all media for
// given remote account is loaded where possible.
func (p *Processor) LoadAccountMedia(
	ctx context.Context,
	requester *gtsmodel.Account,
	account *gtsmodel.Account,
) {
	if requester == nil {
		return
	}
	p.loadAccountMedia(ctx, requester, account)
}

// LoadStatusMedia ensures that all media for
// given remote status is loaded where possible.
func (p *Processor) LoadStatusMedia(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) {
	if requester == nil {
		return
	}
	p.loadAccountMedia(ctx, requester, status.Account)
	if status.BoostOfAccount != nil {
		p.loadAccountMedia(ctx, requester, status.BoostOfAccount)
	}
	p.loadStatusMedia(ctx, requester, status)
	if status.BoostOf != nil {
		p.loadStatusMedia(ctx, requester, status.BoostOf)
	}
}

// loadAccountMedia contains the
// "meat" of LoadAccountMedia().
func (p *Processor) loadAccountMedia(
	ctx context.Context,
	requester *gtsmodel.Account,
	account *gtsmodel.Account,
) {
	if account.IsLocal() {
		return
	}

	var err error

	if account.HeaderMediaAttachment != nil {
		// Ensure account header attachment is loaded and cached.
		//
		// If media attachment is still processing, this call will block.
		account.HeaderMediaAttachment, err = p.federator.RefreshMedia(ctx,
			requester.Username,
			account.HeaderMediaAttachment,
			media.AdditionalMediaInfo{},
			false, // force
			false, // async
		)
		if err != nil {
			log.Errorf(ctx, "error refreshing header attachment %s: %v", account.HeaderMediaAttachment.RemoteURL, err)
		}
	}

	if account.AvatarMediaAttachment != nil {
		// Ensure account avatar attachment is loaded and cached.
		//
		// If media attachment is still processing, this call will block.
		account.AvatarMediaAttachment, err = p.federator.RefreshMedia(ctx,
			requester.Username,
			account.AvatarMediaAttachment,
			media.AdditionalMediaInfo{},
			false, // force
			false, // async
		)
		if err != nil {
			log.Errorf(ctx, "error refreshing avatar attachment %s: %v", account.AvatarMediaAttachment.RemoteURL, err)
		}
	}
}

// loadStatusMedia contains the
// "meat" of LoadStatusMedia().
func (p *Processor) loadStatusMedia(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) {
	if !status.Flags.Local() {
		return
	}

	// Ensure status media attachments are loaded,
	// the below funcion checks if already cached.
	//
	// If media attachments are already processing
	// from previous dereference, this will block.
	for i, attach := range status.Attachments {
		attach, err := p.federator.RefreshMedia(ctx,
			requester.Username,
			attach,
			media.AdditionalMediaInfo{},
			false, // force
			false, // async
		)
		if err != nil {
			log.Errorf(ctx, "error refreshing media attachment %s: %v", attach.RemoteURL, err)
		}

		// Set media attachment model.
		status.Attachments[i] = attach
	}
}
