package services

import (
	"fmt"
	"github.com/neo4j-graphacademy/neoflix/pkg/fixtures"
	"github.com/neo4j-graphacademy/neoflix/pkg/ioutils"

	"github.com/neo4j-graphacademy/neoflix/pkg/routes/paging"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type Movie = map[string]interface{}

type MovieService interface {
	FindAll(userId string, page *paging.Paging) ([]Movie, error)

	FindAllByGenre(genre, userId string, page *paging.Paging) ([]Movie, error)

	FindAllByActorId(actorId string, userId string, page *paging.Paging) ([]Movie, error)

	FindAllByDirectorId(actorId string, userId string, page *paging.Paging) ([]Movie, error)

	FindOneById(id string, userId string) (Movie, error)

	FindAllBySimilarity(id string, userId string, page *paging.Paging) ([]Movie, error)
}

type neo4jMovieService struct {
	loader *fixtures.FixtureLoader
	driver neo4j.Driver
}

func NewMovieService(loader *fixtures.FixtureLoader, driver neo4j.Driver) MovieService {
	return &neo4jMovieService{loader: loader, driver: driver}
}

// FindAll should return a paginated list of movies ordered by the `sort`
// parameter and limited to the number passed as `limit`.  The `skip` variable should be
// used to skip a certain number of rows.
//
// If a userId value is supplied, a `favorite` boolean property should be returned to
// signify whether the user has added the movie to their "My Favorites" list.
// tag::all[]
func (ms *neo4jMovieService) FindAll(userId string, page *paging.Paging) (_ []Movie, err error) {
	session := ms.driver.NewSession(neo4j.SessionConfig{})

	defer func() {
		err = ioutils.DeferredClose(session, err)
	}()

	results, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		favorites, err := getUserFavorites(tx, userId)
		if err != nil {
			return nil, err
		}

		result, err := tx.Run(fmt.Sprintf(`
			MATCH (m:Movie)
			WHERE m.`+"`%[1]s`"+` IS NOT NULL
			RETURN m { 
				.* ,
				favorite: m.tmdbId IN $favorites
			} AS movie
			ORDER BY m.`+"`%[1]s`"+` %s
			SKIP $skip
			LIMIT $limit
		`, page.Sort(), page.Order()), map[string]interface{}{
			"skip":      page.Skip(),
			"limit":     page.Limit(),
			"favorites": favorites,
		})
		if err != nil {
			return nil, err
		}

		records, err := result.Collect()
		if err != nil {
			return nil, err
		}

		var results []map[string]interface{}
		for _, record := range records {
			movie, _ := record.Get("movie")
			results = append(results, movie.(map[string]interface{}))
		}

		return results, nil
	})

	if err != nil {
		return nil, err
	}

	return fixtures.Slice(results.([]Movie), page.Skip(), page.Limit()), nil
}

// end::all[]

// FindAllByGenre should return a paginated list of movies that have a relationship to the
// supplied Genre.
//
// Results should be ordered by the `sort` parameter, and in the direction specified
// in the `order` parameter.
// Results should be limited to the number passed as `limit`.
// The `skip` variable should be used to skip a certain number of rows.
//
// If a userId value is supplied, a `favorite` boolean property should be returned to
// signify whether the user has added the movie to their "My Favorites" list.
//
// tag::getByGenre[]
func (ms *neo4jMovieService) FindAllByGenre(genre string, userId string, page *paging.Paging) (_ []Movie, err error) {
	session := ms.driver.NewSession(neo4j.SessionConfig{})

	defer func() {
		err = ioutils.DeferredClose(session, err)
	}()

	results, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		favorites, err := getUserFavorites(tx, userId)
		if err != nil {
			return nil, err
		}

		result, err := tx.Run(fmt.Sprintf(`
			MATCH (m:Movie)-[:IN_GENRE]->(:Genre {name: $name})
			WHERE m.`+"`%[1]s`"+` IS NOT NULL
			RETURN m { 
				.* ,
				favorite: m.tmdbId IN $favorites
			} AS movie
			ORDER BY m.`+"`%[1]s`"+` %s
			SKIP $skip
			LIMIT $limit
		`, page.Sort(), page.Order()), map[string]interface{}{
			"skip":      page.Skip(),
			"limit":     page.Limit(),
			"favorites": favorites,
			"name":      genre,
		})
		if err != nil {
			return nil, err
		}

		records, err := result.Collect()
		if err != nil {
			return nil, err
		}

		var results []map[string]interface{}
		for _, record := range records {
			movie, _ := record.Get("movie")
			results = append(results, movie.(map[string]interface{}))
		}

		return results, nil
	})

	if err != nil {
		return nil, err
	}

	return fixtures.Slice(results.([]Movie), page.Skip(), page.Limit()), nil
}

// end::getByGenre[]

// FindAllByActorId should return a paginated list of movies that have an ACTED_IN relationship
// to a Person with the id supplied
//
// Results should be ordered by the `sort` parameter, and in the direction specified
// in the `order` parameter.
// Results should be limited to the number passed as `limit`.
// The `skip` variable should be used to skip a certain number of rows.
//
// If a userId value is supplied, a `favorite` boolean property should be returned to
// signify whether the user has added the movie to their "My Favorites" list.
// tag::getForActor[]
func (ms *neo4jMovieService) FindAllByActorId(actorId string, userId string, page *paging.Paging) (_ []Movie, err error) {
	session := ms.driver.NewSession(neo4j.SessionConfig{})

	defer func() {
		err = ioutils.DeferredClose(session, err)
	}()

	results, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		favorites, err := getUserFavorites(tx, userId)
		if err != nil {
			return nil, err
		}

		result, err := tx.Run(fmt.Sprintf(`
			MATCH (:Person {tmdbId: $id})-[:ACTED_IN]->(m:Movie)
			WHERE m.`+"`%[1]s`"+` IS NOT NULL
			RETURN m { 
				.* ,
				favorite: m.tmdbId IN $favorites
			} AS movie
			ORDER BY m.`+"`%[1]s`"+` %s
			SKIP $skip
			LIMIT $limit
		`, page.Sort(), page.Order()), map[string]interface{}{
			"skip":      page.Skip(),
			"limit":     page.Limit(),
			"favorites": favorites,
			"id":        actorId,
		})
		if err != nil {
			return nil, err
		}

		records, err := result.Collect()
		if err != nil {
			return nil, err
		}

		var results []map[string]interface{}
		for _, record := range records {
			movie, _ := record.Get("movie")
			results = append(results, movie.(map[string]interface{}))
		}

		return results, nil
	})

	if err != nil {
		return nil, err
	}

	return fixtures.Slice(results.([]Movie), page.Skip(), page.Limit()), nil
}

// end::getForActor[]

// FindAllByDirectorId should return a paginated list of movies that have an DIRECTED relationship
// to a Person with the id supplied
//
// Results should be ordered by the `sort` parameter, and in the direction specified
// in the `order` parameter.
// Results should be limited to the number passed as `limit`.
// The `skip` variable should be used to skip a certain number of rows.
//
// If a userId value is supplied, a `favorite` boolean property should be returned to
// signify whether the user has added the movie to their "My Favorites" list.
// tag::getForDirector[]
func (ms *neo4jMovieService) FindAllByDirectorId(actorId string, userId string, page *paging.Paging) (_ []Movie, err error) {
	session := ms.driver.NewSession(neo4j.SessionConfig{})

	defer func() {
		err = ioutils.DeferredClose(session, err)
	}()

	results, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		favorites, err := getUserFavorites(tx, userId)
		if err != nil {
			return nil, err
		}

		result, err := tx.Run(fmt.Sprintf(`
			MATCH (:Person {tmdbId: $id})-[:DIRECTED]->(m:Movie)
			WHERE m.`+"`%[1]s`"+` IS NOT NULL
			RETURN m { 
				.* ,
				favorite: m.tmdbId IN $favorites
			} AS movie
			ORDER BY m.`+"`%[1]s`"+` %s
			SKIP $skip
			LIMIT $limit
		`, page.Sort(), page.Order()), map[string]interface{}{
			"skip":      page.Skip(),
			"limit":     page.Limit(),
			"favorites": favorites,
			"id":        actorId,
		})
		if err != nil {
			return nil, err
		}

		records, err := result.Collect()
		if err != nil {
			return nil, err
		}

		var results []map[string]interface{}
		for _, record := range records {
			movie, _ := record.Get("movie")
			results = append(results, movie.(map[string]interface{}))
		}

		return results, nil
	})

	if err != nil {
		return nil, err
	}

	return fixtures.Slice(results.([]Movie), page.Skip(), page.Limit()), nil
}

// end::getForDirector[]

// FindOneById finds a Movie node with the ID passed as the `id` parameter.
// Along with the returned payload, a list of actors, directors, and genres should
// be included.
// The number of incoming RATED relationships should also be returned as `ratingCount`
//
// If a userId value is supplied, a `favorite` boolean property should be returned to
// signify whether the user has added the movie to their "My Favorites" list.
// tag::findById[]
func (ms *neo4jMovieService) FindOneById(id string, userId string) (_ Movie, err error) {
	session := ms.driver.NewSession(neo4j.SessionConfig{})

	defer func() {
		err = ioutils.DeferredClose(session, err)
	}()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		favorites, err := getUserFavorites(tx, userId)
		if err != nil {
			return nil, err
		}

		result, err := tx.Run(`
			MATCH (m:Movie {tmdbId: $id})
			RETURN m {
			  .*,
				actors: [ (a)-[r:ACTED_IN]->(m) | a { .*, role: r.role } ],
				directors: [ (d)-[:DIRECTED]->(m) | d { .* } ],
				genres: [ (m)-[:IN_GENRE]->(g) | g { .name }],
				ratingCount: size((m)<-[:RATED]-()),
				favorite: m.tmdbId IN $favorites
			} AS movie
			LIMIT 1`,
			map[string]interface{}{
				"id":        id,
				"favorites": favorites,
			})
		if err != nil {
			return nil, err
		}

		record, err := result.Single()
		if err != nil {
			return nil, err
		}
		movie, _ := record.Get("movie")
		return movie, nil
	})

	if err != nil {
		return nil, err
	}
	return result.(Movie), nil
}

// end::findById[]

// FindAllBySimilarity should return a paginated list of similar movies to the Movie with the
// id supplied.  This similarity is calculated by finding movies that have many first
// degree connections in common: Actors, Directors and Genres.
//
// Results should be ordered by the `sort` parameter, and in the direction specified
// in the `order` parameter.
// Results should be limited to the number passed as `limit`.
// The `skip` variable should be used to skip a certain number of rows.
//
// If a userId value is supplied, a `favorite` boolean property should be returned to
// signify whether the user has added the movie to their "My Favorites" list.
// tag::getSimilarMovies[]
func (ms *neo4jMovieService) FindAllBySimilarity(id string, userId string, page *paging.Paging) (_ []Movie, err error) {
	session := ms.driver.NewSession(neo4j.SessionConfig{})
	defer func() {
		err = ioutils.DeferredClose(session, err)
	}()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		favorites, err := getUserFavorites(tx, userId)
		if err != nil {
			return nil, err
		}

		// Doesn't work in v5
		result, err := tx.Run(`
			MATCH (:Movie {tmdbId: $id})-[:IN_GENRE|ACTED_IN|DIRECTED]->()<-[:IN_GENRE|ACTED_IN|DIRECTED]-(m)
			WHERE m.imdbRating IS NOT NULL

			WITH m, count(*) AS inCommon
			WITH m, inCommon, m.imdbRating * inCommon AS score
			ORDER BY score DESC

			SKIP $skip
			LIMIT $limit

			RETURN m {
				.*,
				score: score,
				favorite: m.tmdbId IN $favorites
			} AS movie
		`, map[string]interface{}{
			"id":        id,
			"favorites": favorites,
			"skip":      page.Skip(),
			"limit":     page.Limit(),
		})
		if err != nil {
			return nil, err
		}

		records, err := result.Collect()
		if err != nil {
			return nil, err
		}

		var results []map[string]interface{}
		for _, record := range records {
			movie, _ := record.Get("movie")
			results = append(results, movie.(map[string]interface{}))
		}

		return results, nil
	})

	if err != nil {
		return nil, err
	}
	return result.([]Movie), nil
}

// end::getSimilarMovies[]

// getUserFavorites should return a list of tmdbId properties for the movies that
// the user has added to their 'My Favorites' list.
// tag::getUserFavorites[]
func getUserFavorites(tx neo4j.Transaction, userId string) ([]string, error) {
	if userId == "" {
		return nil, nil
	}

	result, err := tx.Run(`
		MATCH (u:User {userId: $userId})-[:HAS_FAVORITE]->(m)
		RETURN m.tmdbId AS id
	`, map[string]interface{}{"userId": userId})
	if err != nil {
		return nil, err
	}

	var ids []string
	for result.Next() {
		record := result.Record()
		id, _ := record.Get("id")
		ids = append(ids, id.(string))
	}
	return ids, nil
}

// end::getUserFavorites[]
