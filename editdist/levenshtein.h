#ifndef LEVENSHTEIN_H_
#define LEVENSHTEIN_H_

#include <map>
#include <string>
#include <vector>

namespace cos981 {

enum EditOp {add, del, sub, equ};
typedef std::pair<char, EditOp> Edit;
std::map<EditOp, int> stats(std::vector<Edit> edits);
std::string to_string(EditOp& eo);

std::pair<int, std::vector<Edit> > levenshtein(std::string a, std::string b);

} // namespace cos981

#endif // LEVENSHTEIN_H_
