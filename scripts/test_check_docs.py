"""Regression checks for standalone CI and full-workspace link resolution."""
from pathlib import Path
import tempfile
import unittest

from check_docs import check_documents


class DocumentationLinksTest(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.directory = Path(self.temporary.name)
        self.root = self.directory / "checkout" / "arbitrary-name"
        self.root.mkdir(parents=True)
        self.home = self.directory / "runner-home"
        self.home.mkdir()

    def document(self, content):
        (self.root / "README.md").write_text(content)

    def test_relocated_checkout_with_unavailable_external_repository(self):
        (self.root / "docs").mkdir()
        (self.root / "docs" / "usage guide.md").write_text("guide")
        self.document(
            "[guide](<~/codespace/magicOrm/docs/usage%20guide.md:12#section>)\n"
            "[root](~/codespace/magicOrm)\n"
            "[platform](~/dataspace/magicPlatform/docs/rules.md)\n"
            "[web](https://example.test) [anchor](#section)"
        )
        result = check_documents(self.root, self.home)
        self.assertEqual(result.failures, [])
        self.assertEqual(result.checked, 2)
        self.assertEqual(result.unavailable, {"~/dataspace/magicPlatform": 1})

    def test_missing_checkout_file_cannot_fall_back_to_home(self):
        canonical = self.home / "codespace" / "magicOrm"
        canonical.mkdir(parents=True)
        (canonical / "missing.md").write_text("stale copy")
        self.document("[missing](~/codespace/magicOrm/missing.md)")
        result = check_documents(self.root, self.home)
        self.assertEqual(len(result.failures), 1)
        self.assertIn("missing file", result.failures[0])

    def test_available_external_repository_still_rejects_broken_links(self):
        external = self.home / "dataspace" / "magicPlatform"
        external.mkdir(parents=True)
        (external / "present.md").write_text("present")
        self.document(
            "[present](~/dataspace/magicPlatform/present.md)\n"
            "[missing](~/dataspace/magicPlatform/missing.md)"
        )
        result = check_documents(self.root, self.home)
        self.assertEqual(result.checked, 2)
        self.assertEqual(result.unavailable, {})
        self.assertEqual(len(result.failures), 1)

    def test_sibling_checkout_outside_home(self):
        sibling = self.root.parent / "magicCommon"
        sibling.mkdir()
        (sibling / "README.md").write_text("common")
        self.document("[common](~/codespace/magicCommon/README.md)")
        result = check_documents(self.root, self.home)
        self.assertEqual(result.checked, 1)
        self.assertEqual(result.failures, [])
        self.assertEqual(result.unavailable, {})

    def test_invalid_paths_remain_errors_and_vendor_is_excluded(self):
        (self.root / "vendor").mkdir()
        (self.root / "vendor" / "README.md").write_text("[bad](missing.md)")
        self.document(
            "[relative](../../missing.md)\n"
            "[escape](~/codespace/magicOrm/%2e%2e/missing.md)"
        )
        result = check_documents(self.root, self.home)
        self.assertEqual(len(result.failures), 2)
        self.assertEqual(result.unavailable, {})


if __name__ == "__main__":
    unittest.main()
