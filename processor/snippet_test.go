package processor

import (
	"testing"
)

func TestExtractLocationZero(t *testing.T) {
	loc := extractLocations([]string{"test"}, "test")

	if loc[0] != 0 {
		t.Error("Expected to be found at 0")
	}
}

func TestExtractLocationOne(t *testing.T) {
	loc := extractLocations([]string{"test"}, " test")

	if loc[0] != 1 {
		t.Error("Expected to be found at 1")
	}
}

func TestExtractLocationsMultiple(t *testing.T) {
	sampleText := `Welcome to Yahoo!, the world’s most visited home page. Quickly find what you’re searching for, get in touch with friends and stay in-the-know with the latest news and information. CloudSponge provides an interface to easily enable your users to import contacts from a variety of the most popular webmail services including Yahoo, Gmail and Hotmail/MSN as well as popular desktop address books such as Mac Address Book and Outlook.`
	searchTerm := []string{"yahoo", "outlook"}

	extractLocations(searchTerm, sampleText)

	t.Error("something")
}

func TestExtractLocationsNoRegexMultiple(t *testing.T) {
	sampleText := `testtest`
	searchTerm := []string{"test"}

	extractLocationsNoRegex(searchTerm, sampleText)

	t.Error("something")
}

func TestExtractRelevant(t *testing.T) {
	sampleText := `Welcome to Yahoo!, the world’s most visited home page. Quickly find what you’re searching for, get in touch with friends and stay in-the-know with the latest news and information. CloudSponge provides an interface to easily enable your users to import contacts from a variety of the most popular webmail services including Yahoo, Gmail and Hotmail/MSN as well as popular desktop address books such as Mac Address Book and Outlook.`
	searchTerm := []string{"yahoo", "and", "outlook"}

	relevant := extractRelevant(searchTerm, sampleText, 300, 70, "…")

	if relevant != "…including Yahoo, Gmail and Hotmail/MSN as well as popular desktop address books such as Mac Address Book and Outlook." {
		t.Error(relevant)
	}
}

func TestExtractRelevantCode(t *testing.T) {
	sampleText := `package com.boyter.SpellingCorrector;

import java.util.*;
import java.util.stream.Stream;

public class SpellingCorrector implements ISpellingCorrector {
  private Map<String, Integer> dictionary = null;

  public SpellingCorrector(int lruCount) {
      this.dictionary = Collections.synchronizedMap(new LruCache<>(lruCount));
  }

  @Override
  public void putWord(String word) {
      word = word.toLowerCase();
      if (dictionary.containsKey(word)) {
          dictionary.put(word, (dictionary.get(word) + 1));
      }
      else {
          dictionary.put(word, 1);
      }
  }

  @Override
  public String correct(String word) {
      if (word == null || word.trim().isEmpty()) {
          return word;
      }

      word = word.toLowerCase();

      if (dictionary.containsKey(word)) {
          return word;
      }

      Map<String, Integer> possibleMatches = new HashMap<>();

      List<String> closeEdits = wordEdits(word);
      for (String closeEdit: closeEdits) {
          if (dictionary.containsKey(closeEdit)) {
              possibleMatches.put(closeEdit, this.dictionary.get(closeEdit));
          }
      }

      if (!possibleMatches.isEmpty()) {
          Object[] matches = this.sortByValue(possibleMatches).keySet().toArray();

          String bestMatch = "";
          for(Object o: matches) {
              if (o.toString().length() == word.length()) {
                  bestMatch = o.toString();
              }
          }

          if (!bestMatch.trim().isEmpty()) {
              return bestMatch;
          }

          return matches[matches.length - 1].toString();
      }

      List<String> furtherEdits = new ArrayList<>();
      for(String closeEdit: closeEdits) {
          furtherEdits.addAll(this.wordEdits(closeEdit));
      }

      for (String futherEdit: furtherEdits) {
          if (dictionary.containsKey(futherEdit)) {
              possibleMatches.put(futherEdit, this.dictionary.get(futherEdit));
          }
      }

      if (!possibleMatches.isEmpty()) {
          Object[] matches = this.sortByValue(possibleMatches).keySet().toArray();

          String bestMatch = "";
          for(Object o: matches) {
              if (o.toString().length() == word.length()) {
                  bestMatch = o.toString();
              }
          }

          if (!bestMatch.trim().isEmpty()) {
              return bestMatch;
          }

          return matches[matches.length - 1].toString();
      }

      return word;
  }

  @Override
  public boolean containsWord(String word) {
      if (dictionary.containsKey(word)) {
          return true;
      }

      return false;
  }

  private List<String> wordEdits(String word) {
      List<String> closeWords = new ArrayList<String>();

      for (int i = 1; i < word.length() + 1; i++) {
          for (char character = 'a'; character <= 'z'; character++) {
              StringBuilder sb = new StringBuilder(word);
              sb.insert(i, character);
              closeWords.add(sb.toString());
          }
      }

      for (int i = 1; i < word.length(); i++) {
          for (char character = 'a'; character <= 'z'; character++) {
              StringBuilder sb = new StringBuilder(word);
              sb.setCharAt(i, character);
              closeWords.add(sb.toString());

              sb = new StringBuilder(word);
              sb.deleteCharAt(i);
              closeWords.add(sb.toString());
          }
      }

      return closeWords;
  }

  public static <K, V extends Comparable<? super V>> Map<K, V> sortByValue( Map<K, V> map ) {
      Map<K, V> result = new LinkedHashMap<>();
      Stream<Map.Entry<K, V>> st = map.entrySet().stream();

      st.sorted( Map.Entry.comparingByValue() ).forEachOrdered( e -> result.put(e.getKey(), e.getValue()) );

      return result;
  }

  public class LruCache<A, B> extends LinkedHashMap<A, B> {
      private final int maxEntries;

      public LruCache(final int maxEntries) {
          super(maxEntries + 1, 1.0f, true);
          this.maxEntries = maxEntries;
      }

      @Override
      protected boolean removeEldestEntry(final Map.Entry<A, B> eldest) {
          return super.size() > maxEntries;
      }
  }
}
`
	searchTerm := []string{"extends", "linked"}

	relevant := extractRelevant(searchTerm, sampleText, 300, 50, "…")

	t.Error(relevant)

}

////////////////// Benchmarks

func BenchmarkExtractLocationsRegex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		extractLocations([]string{"public"}, `package com.boyter.SpellingCorrector;

import java.util.*;
import java.util.stream.Stream;

public class SpellingCorrector implements ISpellingCorrector {
  private Map<String, Integer> dictionary = null;

  public SpellingCorrector(int lruCount) {
      this.dictionary = Collections.synchronizedMap(new LruCache<>(lruCount));
  }

  @Override
  public void putWord(String word) {
      word = word.toLowerCase();
      if (dictionary.containsKey(word)) {
          dictionary.put(word, (dictionary.get(word) + 1));
      }
      else {
          dictionary.put(word, 1);
      }
  }

  @Override
  public String correct(String word) {
      if (word == null || word.trim().isEmpty()) {
          return word;
      }

      word = word.toLowerCase();

      if (dictionary.containsKey(word)) {
          return word;
      }

      Map<String, Integer> possibleMatches = new HashMap<>();

      List<String> closeEdits = wordEdits(word);
      for (String closeEdit: closeEdits) {
          if (dictionary.containsKey(closeEdit)) {
              possibleMatches.put(closeEdit, this.dictionary.get(closeEdit));
          }
      }

      if (!possibleMatches.isEmpty()) {
          Object[] matches = this.sortByValue(possibleMatches).keySet().toArray();

          String bestMatch = "";
          for(Object o: matches) {
              if (o.toString().length() == word.length()) {
                  bestMatch = o.toString();
              }
          }

          if (!bestMatch.trim().isEmpty()) {
              return bestMatch;
          }

          return matches[matches.length - 1].toString();
      }

      List<String> furtherEdits = new ArrayList<>();
      for(String closeEdit: closeEdits) {
          furtherEdits.addAll(this.wordEdits(closeEdit));
      }

      for (String futherEdit: furtherEdits) {
          if (dictionary.containsKey(futherEdit)) {
              possibleMatches.put(futherEdit, this.dictionary.get(futherEdit));
          }
      }

      if (!possibleMatches.isEmpty()) {
          Object[] matches = this.sortByValue(possibleMatches).keySet().toArray();

          String bestMatch = "";
          for(Object o: matches) {
              if (o.toString().length() == word.length()) {
                  bestMatch = o.toString();
              }
          }

          if (!bestMatch.trim().isEmpty()) {
              return bestMatch;
          }

          return matches[matches.length - 1].toString();
      }

      return word;
  }

  @Override
  public boolean containsWord(String word) {
      if (dictionary.containsKey(word)) {
          return true;
      }

      return false;
  }

  private List<String> wordEdits(String word) {
      List<String> closeWords = new ArrayList<String>();

      for (int i = 1; i < word.length() + 1; i++) {
          for (char character = 'a'; character <= 'z'; character++) {
              StringBuilder sb = new StringBuilder(word);
              sb.insert(i, character);
              closeWords.add(sb.toString());
          }
      }

      for (int i = 1; i < word.length(); i++) {
          for (char character = 'a'; character <= 'z'; character++) {
              StringBuilder sb = new StringBuilder(word);
              sb.setCharAt(i, character);
              closeWords.add(sb.toString());

              sb = new StringBuilder(word);
              sb.deleteCharAt(i);
              closeWords.add(sb.toString());
          }
      }

      return closeWords;
  }

  public static <K, V extends Comparable<? super V>> Map<K, V> sortByValue( Map<K, V> map ) {
      Map<K, V> result = new LinkedHashMap<>();
      Stream<Map.Entry<K, V>> st = map.entrySet().stream();

      st.sorted( Map.Entry.comparingByValue() ).forEachOrdered( e -> result.put(e.getKey(), e.getValue()) );

      return result;
  }

  public class LruCache<A, B> extends LinkedHashMap<A, B> {
      private final int maxEntries;

      public LruCache(final int maxEntries) {
          super(maxEntries + 1, 1.0f, true);
          this.maxEntries = maxEntries;
      }

      @Override
      protected boolean removeEldestEntry(final Map.Entry<A, B> eldest) {
          return super.size() > maxEntries;
      }
  }
}
`)
	}
}

func BenchmarkExtractLocationsNoRegex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		extractLocationsNoRegex([]string{"public"}, `package com.boyter.SpellingCorrector;

import java.util.*;
import java.util.stream.Stream;

public class SpellingCorrector implements ISpellingCorrector {
  private Map<String, Integer> dictionary = null;

  public SpellingCorrector(int lruCount) {
      this.dictionary = Collections.synchronizedMap(new LruCache<>(lruCount));
  }

  @Override
  public void putWord(String word) {
      word = word.toLowerCase();
      if (dictionary.containsKey(word)) {
          dictionary.put(word, (dictionary.get(word) + 1));
      }
      else {
          dictionary.put(word, 1);
      }
  }

  @Override
  public String correct(String word) {
      if (word == null || word.trim().isEmpty()) {
          return word;
      }

      word = word.toLowerCase();

      if (dictionary.containsKey(word)) {
          return word;
      }

      Map<String, Integer> possibleMatches = new HashMap<>();

      List<String> closeEdits = wordEdits(word);
      for (String closeEdit: closeEdits) {
          if (dictionary.containsKey(closeEdit)) {
              possibleMatches.put(closeEdit, this.dictionary.get(closeEdit));
          }
      }

      if (!possibleMatches.isEmpty()) {
          Object[] matches = this.sortByValue(possibleMatches).keySet().toArray();

          String bestMatch = "";
          for(Object o: matches) {
              if (o.toString().length() == word.length()) {
                  bestMatch = o.toString();
              }
          }

          if (!bestMatch.trim().isEmpty()) {
              return bestMatch;
          }

          return matches[matches.length - 1].toString();
      }

      List<String> furtherEdits = new ArrayList<>();
      for(String closeEdit: closeEdits) {
          furtherEdits.addAll(this.wordEdits(closeEdit));
      }

      for (String futherEdit: furtherEdits) {
          if (dictionary.containsKey(futherEdit)) {
              possibleMatches.put(futherEdit, this.dictionary.get(futherEdit));
          }
      }

      if (!possibleMatches.isEmpty()) {
          Object[] matches = this.sortByValue(possibleMatches).keySet().toArray();

          String bestMatch = "";
          for(Object o: matches) {
              if (o.toString().length() == word.length()) {
                  bestMatch = o.toString();
              }
          }

          if (!bestMatch.trim().isEmpty()) {
              return bestMatch;
          }

          return matches[matches.length - 1].toString();
      }

      return word;
  }

  @Override
  public boolean containsWord(String word) {
      if (dictionary.containsKey(word)) {
          return true;
      }

      return false;
  }

  private List<String> wordEdits(String word) {
      List<String> closeWords = new ArrayList<String>();

      for (int i = 1; i < word.length() + 1; i++) {
          for (char character = 'a'; character <= 'z'; character++) {
              StringBuilder sb = new StringBuilder(word);
              sb.insert(i, character);
              closeWords.add(sb.toString());
          }
      }

      for (int i = 1; i < word.length(); i++) {
          for (char character = 'a'; character <= 'z'; character++) {
              StringBuilder sb = new StringBuilder(word);
              sb.setCharAt(i, character);
              closeWords.add(sb.toString());

              sb = new StringBuilder(word);
              sb.deleteCharAt(i);
              closeWords.add(sb.toString());
          }
      }

      return closeWords;
  }

  public static <K, V extends Comparable<? super V>> Map<K, V> sortByValue( Map<K, V> map ) {
      Map<K, V> result = new LinkedHashMap<>();
      Stream<Map.Entry<K, V>> st = map.entrySet().stream();

      st.sorted( Map.Entry.comparingByValue() ).forEachOrdered( e -> result.put(e.getKey(), e.getValue()) );

      return result;
  }

  public class LruCache<A, B> extends LinkedHashMap<A, B> {
      private final int maxEntries;

      public LruCache(final int maxEntries) {
          super(maxEntries + 1, 1.0f, true);
          this.maxEntries = maxEntries;
      }

      @Override
      protected boolean removeEldestEntry(final Map.Entry<A, B> eldest) {
          return super.size() > maxEntries;
      }
  }
}
`)
	}
}
