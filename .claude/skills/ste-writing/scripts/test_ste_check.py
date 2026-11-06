import unittest

from ste_check import check_text


def rules(text):
    return [f.rule for f in check_text(text)]


def words(n):
    return " ".join(["word"] * n) + "."


class SteCheckTest(unittest.TestCase):
    def test_semicolon(self):
        self.assertEqual(rules("Do X; do Y."), ["semicolon"])

    def test_long_sentence(self):
        self.assertEqual(rules(words(26)), ["sentence-length"])
        self.assertEqual(rules(words(20)), [])

    def test_present_perfect(self):
        self.assertEqual(rules("The job has been sent."), ["present-perfect"])

    def test_skips_code(self):
        fenced = "Run this:\n\n```bash\na; b\n```\n"
        self.assertEqual(rules(fenced), [])
        self.assertEqual(rules("Run `a; b` now."), [])

    def test_skips_tables_and_headings(self):
        long_line = " ".join(["word"] * 30)
        self.assertEqual(rules("| " + long_line + " |"), [])
        self.assertEqual(rules("# " + long_line), [])

    def test_splits_after_closing_quote_or_paren(self):
        short = " ".join(["a"] * 15)
        self.assertEqual(rules(f"({short}.) {short}."), [])
        self.assertEqual(rules(f'He said "{short}." {short}.'), [])

    def test_list_marker_is_not_a_word(self):
        item = " ".join(["word"] * 25) + "."
        self.assertEqual(rules("- " + item), [])
        self.assertEqual(rules("1. " + item), [])

    def test_fence_closes_only_on_matching_marker(self):
        self.assertEqual(rules("```markdown\n~~~\na; b\n~~~\n```\n"), [])
        self.assertEqual(rules("````md\n```\na; b\n````\n"), [])

    def test_sentence_wrapped_across_lines(self):
        text = " ".join(["a"] * 15) + "\n" + " ".join(["b"] * 16) + ".\n"
        found = check_text(text)
        self.assertEqual([f.rule for f in found], ["sentence-length"])
        self.assertEqual(found[0].line, 1)

    def test_entities_and_urls_are_not_semicolons(self):
        self.assertEqual(rules("Use &amp; here."), [])
        self.assertEqual(rules("See https://example.com/a;b now."), [])

    def test_invalid_utf8_does_not_stop_the_run(self):
        import io
        import os
        import tempfile
        from contextlib import redirect_stdout

        from ste_check import main

        with tempfile.TemporaryDirectory() as tmp:
            bad = os.path.join(tmp, "bad.md")
            good = os.path.join(tmp, "good.md")
            with open(bad, "wb") as handle:
                handle.write(b"\xff\xfe bad\n")
            with open(good, "w", encoding="utf-8") as handle:
                handle.write("Do X; do Y.\n")
            out = io.StringIO()
            with redirect_stdout(out):
                code = main([bad, good])
        self.assertEqual(code, 0)
        self.assertIn("read-error", out.getvalue())
        self.assertIn("semicolon", out.getvalue())


if __name__ == "__main__":
    unittest.main()
