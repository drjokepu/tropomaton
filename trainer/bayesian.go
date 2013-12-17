package main

import "database/sql"
import "fmt"
import "github.com/jbrukh/bayesian"
import "os"
import "strings"
import "syscall"

const (
	trope    bayesian.Class = "trope"
	notTrope bayesian.Class = "not_trope"
	work     bayesian.Class = "work"
	notWork  bayesian.Class = "not_work"
)

const (
	tropeIndex    = 0
	notTropeIndex = 1
	workIndex     = 2
	notWorkIndex  = 3
)

const (
	tropeClass = 0
	workClass  = 1
	otherClass = 2
)

const classifierFilename = "classifier.dat"

func trainPage(pageId, class int) error {
	err := run(func(tx *sql.Tx) error {
		page, err := getPage(pageId, tx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to get page from database.")
			return err
		}

		err = page.updateHumanClass(class, tx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to update page class in database.")
			return err
		}

		err = acquireClassifierLock()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to acquire exclusive lock to classifier data file.")
			return err
		}

		defer releaseClassifierLock()

		var classifier *bayesian.Classifier
		isEmpty, err := isClassifierEmpty()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to determine if classifier is empty.")
			return err
		}

		if isEmpty {
			classifier = bayesian.NewClassifier(trope, notTrope, work, notWork)
		} else {
			classifier, err = loadClassifier()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to load classifier.")
				return err
			}
		}

		bayesianClasses := getBayesianClasses(class)
		learnInClasses(getDocument(page.text), bayesianClasses, classifier)

		err = saveClassifier(classifier)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save classifier to data file.")
			return err
		}

		fmt.Println(pageId, page.url, page.title, "---", class, bayesianClasses)

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func guessPageClass(pageId int) (int, error) {
	var guessedClass int
	err := run(func(tx *sql.Tx) error {
		page, err := getPage(pageId, tx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to get page from database.")
			return err
		}

		err = acquireClassifierLock()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to acquire exclusive lock to classifier data file.")
			return err
		}

		defer releaseClassifierLock()

		classifier, err := loadClassifier()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to load classifier.")
			return err
		}

		guessedClass = guessDocumentClass(getDocument(page.text), classifier)

		err = saveClassifier(classifier)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save classifier to data file.")
			return err
		}

		return nil
	})

	if err != nil {
		return -1, err
	}

	return guessedClass, nil
}

func openClassifier() (int, error) {
	fd, err := syscall.Open(classifierFilename,
		syscall.O_RDONLY|syscall.O_CREAT,
		syscall.S_IRUSR|syscall.S_IWUSR|syscall.S_IRGRP|syscall.S_IROTH)
	return fd, err
}

func acquireClassifierLock() error {
	fd, err := openClassifier()
	if err != nil {
		return err
	}

	defer syscall.Close(fd)

	err = syscall.Flock(fd, syscall.LOCK_EX)
	return err
}

func releaseClassifierLock() error {
	fd, err := openClassifier()
	if err != nil {
		return err
	}

	defer syscall.Close(fd)

	err = syscall.Flock(fd, syscall.LOCK_UN)
	return err
}

func isClassifierEmpty() (bool, error) {
	fd, err := openClassifier()
	if err != nil {
		return false, err
	}

	defer syscall.Close(fd)

	var stat syscall.Stat_t
	err = syscall.Fstat(fd, &stat)
	if err != nil {
		return false, err
	}

	return stat.Size == 0, nil
}

func loadClassifier() (*bayesian.Classifier, error) {
	return bayesian.NewClassifierFromFile(classifierFilename)
}

func saveClassifier(classifier *bayesian.Classifier) error {
	return classifier.WriteToFile(classifierFilename)
}

func learnInClasses(doc []string, classes []bayesian.Class, classifier *bayesian.Classifier) {
	for _, class := range classes {
		learnInClass(doc, class, classifier)
	}
}

func learnInClass(doc []string, class bayesian.Class, classifier *bayesian.Classifier) {
	classifier.Learn(doc, class)
}

func getBayesianClasses(class int) []bayesian.Class {
	switch class {
	case tropeClass:
		return []bayesian.Class{trope, notWork}
	case workClass:
		return []bayesian.Class{notTrope, work}
	default:
		return []bayesian.Class{notTrope, notWork}
	}
}

func getClassName(class int) string {
	switch class {
	case tropeClass:
		return "trope"
	case workClass:
		return "work"
	default:
		return "other"
	}
}

func getDocument(text string) []string {
	return strings.Split(text, " ")
}

func guessDocumentClass(doc []string, classifier *bayesian.Classifier) int {
	scores, _, _ := classifier.LogScores(doc)
	isTrope := scores[tropeIndex] > scores[notTropeIndex]
	isWork := scores[workIndex] > scores[notWorkIndex]

	if isTrope {
		if isWork {
			if scores[tropeIndex] >= scores[workIndex] {
				return tropeClass
			} else {
				return workClass
			}
		} else {
			return tropeClass
		}
	} else {
		if isWork {
			return workClass
		} else {
			return otherClass
		}
	}
}
