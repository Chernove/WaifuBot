package db

import (
	"context"
	"strings"
	"time"

	"github.com/Karitham/WaifuBot/internal/discord"
	"github.com/Karitham/corde"
	"github.com/Masterminds/squirrel"
)

var _ discord.Store = (*Queries)(nil)

// PutChar a char in the database
func (q *Queries) PutChar(ctx context.Context, userID corde.Snowflake, c discord.Character) error {
	if q.tx == nil {
		return q.asTx(func(q *Queries) error {
			return q.PutChar(ctx, userID, c)
		})
	}

	p := insertCharParams{
		ID:     c.ID,
		UserID: uint64(c.UserID),
		Image:  c.Image,
		Name:   strings.Join(strings.Fields(c.Name), " "),
		Type:   c.Type,
	}

	// In case user doesn't exist, query will fail
	if err := q.insertChar(ctx, p); err == nil {
		return nil
	}

	if err := q.createUser(ctx, uint64(userID)); err != nil {
		return err
	}

	// This is a dumb retry mechanism, but should cover our use case
	return q.insertChar(ctx, p)
}

// Chars returns the user's characters
func (q *Queries) Chars(ctx context.Context, userID corde.Snowflake) ([]discord.Character, error) {
	dbchs, err := q.getChars(ctx, uint64(userID))
	if err != nil {
		return nil, err
	}

	chars := make([]discord.Character, 0, len(dbchs))
	for _, c := range dbchs {
		chars = append(chars, discord.Character{
			Date:   c.Date,
			Image:  c.Image,
			Name:   c.Name,
			Type:   c.Type,
			UserID: corde.Snowflake(c.UserID),
			ID:     c.ID,
		})
	}

	return chars, nil
}

// CharsIDs returns the user's character's ID
func (q *Queries) CharsIDs(ctx context.Context, userID corde.Snowflake) ([]int64, error) {
	return q.getCharsID(ctx, uint64(userID))
}

// User returns a user
func (q *Queries) User(ctx context.Context, userID corde.Snowflake) (discord.User, error) {
	dbuser, err := q.getUser(ctx, uint64(userID))
	if err != nil {
		return discord.User{}, err
	}

	return discord.User{
		Date:     dbuser.Date,
		Quote:    dbuser.Quote,
		UserID:   corde.Snowflake(dbuser.UserID),
		Favorite: uint64(dbuser.Favorite.Int64),
	}, nil
}

// updateUser updates a user's properties
func (q *Queries) updateUser(ctx context.Context, userID corde.Snowflake, opts ...func(*squirrel.UpdateBuilder)) error {
	builder := squirrel.Update("users").Where(squirrel.Eq{
		"user_id": userID,
	}).PlaceholderFormat(squirrel.Dollar)

	for _, opt := range opts {
		opt(&builder)
	}

	str, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	if _, err := q.exec(ctx, nil, str, args...); err != nil {
		return err
	}

	return nil
}

// withFavorite sets user favorite
func withFav(f int64) func(*squirrel.UpdateBuilder) {
	return func(s *squirrel.UpdateBuilder) {
		*s = s.Set("favorite", f)
	}
}

// withQuote sets user quote
func withQuote(q string) func(*squirrel.UpdateBuilder) {
	return func(s *squirrel.UpdateBuilder) {
		*s = s.Set("quote", q)
	}
}

// withDate sets the date
func withDate(d time.Time) func(*squirrel.UpdateBuilder) {
	return func(s *squirrel.UpdateBuilder) {
		*s = s.Set("date", d.UTC())
	}
}

// withToken sets the token
func withToken(t string) func(*squirrel.UpdateBuilder) {
	return func(s *squirrel.UpdateBuilder) {
		*s = s.Set("token", t)
	}
}

// SetUserDate sets the user's date
func (q *Queries) SetUserDate(ctx context.Context, userID corde.Snowflake, d time.Time) error {
	return q.updateUser(ctx, userID, withDate(d))
}

// SetUserToken sets the user's token
func (q *Queries) SetUserToken(ctx context.Context, userID corde.Snowflake, token string) error {
	return q.updateUser(ctx, userID, withToken(token))
}

// SetUserFavorite sets the user's favorite
func (q *Queries) SetUserFavorite(ctx context.Context, userID corde.Snowflake, c int64) error {
	return q.updateUser(ctx, userID, withFav(c))
}

// SetUserQuote sets the user's quote
func (q *Queries) SetUserQuote(ctx context.Context, userID corde.Snowflake, quote string) error {
	return q.updateUser(ctx, userID, withQuote(quote))
}

// CharsStartingWith returns characters starting with the given string
func (q *Queries) CharsStartingWith(ctx context.Context, userID corde.Snowflake, s string) ([]discord.Character, error) {
	dbchs, err := q.getCharsWhoseIDStartWith(ctx, getCharsWhoseIDStartWithParams{
		UserID:  uint64(userID),
		Lim:     50,
		Off:     0,
		LikeStr: s + "%",
	})
	if err != nil {
		return nil, err
	}

	chars := make([]discord.Character, 0, len(dbchs))
	for _, c := range dbchs {
		chars = append(chars, discord.Character{
			Date:   c.Date,
			Image:  c.Image,
			Name:   c.Name,
			Type:   c.Type,
			UserID: corde.Snowflake(c.UserID),
			ID:     c.ID,
		})
	}

	return chars, nil
}

// Profile returns the user's profile
func (q *Queries) Profile(ctx context.Context, userID corde.Snowflake) (discord.Profile, error) {
	p, err := q.getProfile(ctx, uint64(userID))
	if err != nil {
		return discord.Profile{}, err
	}

	return discord.Profile{
		User: discord.User{
			Date:   p.UserDate,
			Quote:  p.UserQuote,
			UserID: corde.Snowflake(p.UserID),
		},
		CharacterCount: int(p.Count),
		Favorite: discord.Character{
			Image:  p.FavoriteImage.String,
			Name:   p.FavoriteName.String,
			UserID: userID,
			ID:     p.FavoriteID.Int64,
		},
	}, nil
}

func (q *Queries) GiveUserChar(ctx context.Context, dst corde.Snowflake, src corde.Snowflake, charID int64) error {
	_, err := q.giveChar(ctx, giveCharParams{
		Given: int64(dst),
		ID:    charID,
		Giver: int64(src),
	})
	return err
}