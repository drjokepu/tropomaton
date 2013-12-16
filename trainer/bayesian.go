package main

import "database/sql"
import "fmt"
import "github.com/jbrukh/bayesian"
import "strings"
import "syscall"

const (
	trope    bayesian.Class = "trope"
	notTrope bayesian.Class = "not_trope"
	work     bayesian.Class = "work"
	notWork  bayesian.Class = "not_work"
)

const classifierFilename = "classifier.dat"

func train(pageId, class int) error {
	fmt.Println("Page:", pageId, "Class:", class)

	err := run(func(tx *sql.Tx) error {
		page, err := getPage(pageId, tx)
		if err != nil {
			return err
		}
		page = page

		err = acquireClassifierLock()
		if err != nil {
			return err
		}

		defer releaseClassifierLock()

		classifier, err := loadClassifier()
		if err != nil {
			return err
		}

		learnInClasses(getDocument(page.text), getBayesianClasses(class), classifier)

		err = saveClassifier(classifier)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
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
	case 0:
		return []bayesian.Class{trope, notWork}
	case 1:
		return []bayesian.Class{notTrope, work}
	default:
		return []bayesian.Class{notTrope, notWork}
	}
}

func getDocument(text string) []string {
	return strings.Split(text, " ")
}
