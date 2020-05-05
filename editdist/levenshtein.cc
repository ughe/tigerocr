#include "levenshtein.h"

#include <algorithm>
#include <cerrno>
#include <cstdlib>
#include <ctime>
#include <fstream>
#include <iomanip>
#include <iostream>
#include <sstream>
#include <stdexcept>

namespace cos981 {

namespace {

template <typename T, typename S>
T depth_first_search(T start, std::pair<bool, std::vector<T> > (*step)(T, S),
                     S state) {
  // Returns start on failure. Otherwise returns the first path found in the
  // depth-first search. step begins with start and returns false and the next
  // possible search objects. If step returns true, then a result has been found
  std::vector<T> stack;
  stack.push_back(start);
  while (stack.size() > 0) {
    T head = stack.back();
    stack.pop_back();
    // Calculate the next states
    std::pair<bool, std::vector<T> > step_result = step(head, state);
    if (step_result.first) {
      return step_result.second[0];
    }
    std::vector<T> next_states = step_result.second;
    // Add next states to the stack
    for (int i = 0; i < next_states.size(); ++i) {
      stack.push_back(next_states[i]);
    }
  }
  return start;
}

struct StateRef {
  std::string& a;
  std::string& b;
  std::vector<std::vector<int> >& cost;
  StateRef(std::string& a, std::string& b, std::vector<std::vector<int> >& cost)
    : a(a), b(b), cost(cost) {}
};

typedef std::pair<int, int> Move;
typedef std::vector<Move> Path;

typedef std::pair<Move, Edit> DecoratedMove;
typedef std::vector<DecoratedMove> DecoratedPath;

std::string to_string(DecoratedMove& dm) {
  std::stringstream buf;
  buf << "[" << to_string(dm.second.second) << ", '" << dm.second.first;
  buf << "', " << "(" << dm.first.first << ", " << dm.first.second << ")]";
  return buf.str();
}

std::string to_string(DecoratedPath& dp) {
  std::stringstream buf;
  std::string sep;
  for (int i = 0; i < dp.size(); ++i) {
    buf << sep << to_string(dp[i]);
    sep = ", ";
  }
  buf << std::endl;
  return buf.str();
}

typedef std::pair<std::pair<int, int>, Move> WeightedMove;

bool cmp_weighted_moves(WeightedMove left, WeightedMove right) {
  std::pair<int, int> lweights = left.first;
  std::pair<int, int> rweights = right.first;
  if (lweights.first == rweights.first) {
    return lweights.second < rweights.second; // diag distance
  } else {
    return lweights.first < rweights.first; // cost
  }
}

std::pair<bool, std::vector<Path> > find_next_paths(Path path, StateRef state) {
  // Returns either true and a vector with the complete path or false and a
  // vector with the partial paths
  int i = path.back().first;
  int j = path.back().second;
  // Unpack the state
  std::string const& a = state.a;
  std::string const& b = state.b;
  std::vector<std::vector<int> >& cost = state.cost;

  // Find all possible moves
  std::vector<WeightedMove> weighted_moves;
  if (i+1 < a.size()+1 && j+1 < b.size()+1) {
    weighted_moves.push_back(std::make_pair(
      std::make_pair(cost[i+1][j+1], std::abs(i-j)),
      Move(i+1, j+1)));
  }
  if (i+1 < a.size()+1) {
    weighted_moves.push_back(std::make_pair(
      std::make_pair(cost[i+1][j], std::abs(i+1-j)),
      Move(i+1, j)));
  }
  if (j+1 < b.size()+1) {
    weighted_moves.push_back(std::make_pair(
      std::make_pair(cost[i][j+1], std::abs(j+1-i)),
      Move(i, j+1)));
  }
  // Sort the moves by cost and then by distance to the diagonal
  std::sort(weighted_moves.begin(), weighted_moves.end(), cmp_weighted_moves);
  // Extract moves from weighted_moves and create new path
  std::vector<Path> paths;
  for (int k = weighted_moves.size()-1; k >= 0; --k) {
    Path new_path = path;
    Move move = weighted_moves[k].second;
    new_path.push_back(move);
    if (move.first == a.size() && move.second == b.size()) {
      // Return if a complete path is found
      std::vector<Path> singleton_path;
      singleton_path.push_back(new_path);
      return std::make_pair(true, singleton_path);
    }
    paths.push_back(new_path);
  }
  // Return all possible partial paths (not complete)
  return std::make_pair(false, paths);
}

DecoratedPath decorate_path(std::string& a, std::string& b,
  std::vector<std::vector<int> >& cost, Path& path) {
  // Return decorated path with edit information for each step
  // First element in path must be Move(0, 0)
  if (path.size() <= 1) {
    return std::vector<DecoratedMove>();
  }
  DecoratedPath dp;
  for (int k = 1; k < path.size(); ++k) {
    int i1 = path[k-1].first;
    int j1 = path[k-1].second;
    int i2 = path[k].first;
    int j2 = path[k].second;
    int c = cost[i2][j2] - cost[i1][j1];
    if (i2==i1+1 && j2==j1+1) {
      if (c == 0) {
        dp.push_back(DecoratedMove(Move(i2, j2), Edit(b[j1], equ)));
      } else {
        dp.push_back(DecoratedMove(Move(i2, j2), Edit(b[j1], sub)));
      }
    } else if (i2==i1+1 && j2==j1) {
      dp.push_back(DecoratedMove(Move(i2, j2), Edit(a[i1], del)));
    } else if (i2==i1 && j2==j1+1) {
      dp.push_back(DecoratedMove(Move(i2, j2), Edit(b[j1], add)));
    } else {
      throw std::invalid_argument("Expected move right, down, or diagonal.");
    }
  }
  return dp;
}

DecoratedPath decorated_path_remove_corners(DecoratedPath dp) {
  // Remove redundant steps in the path recursively
  for (int k = 1; k < dp.size(); ++k) {
    DecoratedMove dm1 = dp[k-1];
    DecoratedMove dm2 = dp[k];
    EditOp dm1_op = dm1.second.second;
    EditOp dm2_op = dm2.second.second;
    if ((dm1_op==add && dm2_op==del) || (dm1_op==del && dm2_op==add)) {
      // Replace two steps with one step
      char c = dm1_op==add ? dm1.second.first : dm2.second.first;
      DecoratedPath modified;
      modified.reserve(dp.size()-1);
      for (int n = 0; n < dp.size(); ++n) {
        if (n < k-1 || n > k) {
          modified.push_back(dp[n]);
        } else if (n == k) {
          modified.push_back(DecoratedMove(
            Move(dm2.first.first, dm2.first.second), Edit(c, sub)));
        }
      }
      return decorated_path_remove_corners(modified);
    }
  }
  return dp;
}

void create_levenshtein_table(std::string& a, std::string& b,
                              std::vector<std::vector<int> >& cost) {
  // Creates distance matrix of cost of transforming from one string to another.
  // Moving right inserts character b[j]. Moving down deletes character a[i-1].
  // Moving diagonally substitutes b[j] for a[i] if a[i] != b[j]. Every cell has
  // 3 possible antecedents. Always choses the lowest cost.
  for (int i = 0; i < a.size()+1; ++i) {
    for (int j = 0; j < b.size()+1; ++j) {
      if (i == 0) {
        cost[i][j] = j; // first row: range 0 ... b.size()
      } else if (j == 0) {
        cost[i][j] = i; // first column: range 0 ... a.size()
      } else {
        cost[i][j] = std::min(std::min(cost[i][j-1]+1, cost[i-1][j]+1),
          cost[i-1][j-1]+(a[i-1] != b[j-1] ? 1 : 0));
      }
    }
  }
}

std::string to_string(std::string& a, std::string& b,
                      std::vector<std::vector<int> >& cost){
  // Returns string representation of a levenshtein table
  int w = 2; // Width of representation
  std::vector<char> aheader(a.begin(), a.end());
  std::vector<char> bheader(b.begin(), b.end());
  std::stringstream buf;
  for (int i = 0; i < a.size()+1; ++i) {
    if (i == 0) {
      buf << std::setfill(' ') << std::setw(2*w) << " " << "  ";
      for (int k = 0; k < bheader.size(); ++k) {
        buf << std::setfill(' ') << std::setw(w) << bheader[k] << " ";
      }
      buf << std::endl;
    }
    for (int j = 0; j < b.size()+1; ++j) {
      if (j == 0) {
        if (i > 0) {
          buf << std::setfill(' ') << std::setw(w) << aheader[i-1] << " ";
        } else {
          buf << std::setfill(' ') << std::setw(w) << " " << " ";
        }
      }
      buf << std::setfill(' ') << std::setw(w) << cost[i][j] << " ";
    }
    buf << std::endl;
  }
  return buf.str();
}

std::map<EditOp, int> stats(DecoratedPath& dp) {
  // Count the number of EditOp uses in a decorated path
  std::map<EditOp, int>stats;
  stats[add] = 0;
  stats[del] = 0;
  stats[sub] = 0;
  stats[equ] = 0;
  for (int i = 0; i < dp.size(); i++) {
    stats[dp[i].second.second] += 1;
  }
  return stats;
}
} // namespace

std::map<EditOp, int> stats(std::vector<Edit> edits) {
  // Count the number of EditOp uses in a decorated path
  std::map<EditOp, int>stats;
  stats[add] = 0;
  stats[del] = 0;
  stats[sub] = 0;
  stats[equ] = 0;
  for (int i = 0; i < edits.size(); i++) {
    stats[edits[i].second] += 1;
  }
  return stats;
}

std::string to_string(std::map<EditOp, int> s) {
  std::stringstream buf;
  std::string sep;
  for (std::map<EditOp, int>::iterator i = s.begin(); i != s.end(); i++) {
    EditOp eo = i->first;
    buf << sep << to_string(eo) << ":" << i->second;
    sep = ", ";
  }
  return buf.str();
}

std::string to_string(EditOp& eo) {
  switch (eo) {
    case add: return "a";
    case del: return "d";
    case sub: return "s";
    case equ: return "e";
    default: return "?";
  }
}

std::pair<int, std::vector<Edit> > levenshtein(std::string a, std::string b) {
  // Returns the Levenshtein distance between `a` and `b` and sequence of edits
  std::vector<std::vector<int> > cost(a.size()+1, std::vector<int>(b.size()+1));

  // Calculate table
  create_levenshtein_table(a, b, cost);

  // Finds a short path with DFS from cost[0][0] to cost[a.size()][b.size()]
  Path start;
  start.push_back(Move(0, 0));
  Path path = depth_first_search(start, find_next_paths, StateRef(a, b, cost));

  // Decorated path
  DecoratedPath dp = decorate_path(a, b, cost, path);

  // Remove corners
  DecoratedPath odp = decorated_path_remove_corners(dp);

  // Edit distance
  std::map<EditOp, int> s = stats(odp);

  // Extract Levenshtein Distance
  int dist = cost[a.size()][b.size()];
  std::vector<Edit> edits;
  for (int i = 0; i < odp.size(); ++i) {
    edits.push_back(odp[i].second);
  }

  // // Debug
  // std::cout << to_string(a, b, cost) << std::endl;
  // std::cout << to_string(dp) << std::endl;
  // std::cout << to_string(odp) << std::endl;
  // std::cout << to_string(s) << std::endl << std::endl;

  return std::make_pair(dist, edits);
}

} // namespace cos981

std::string read(char *filename) {
  // http://insanecoding.blogspot.com/2011/11/how-to-read-in-file-in-c.html
  std::ifstream ifs(filename, std::ios::in);
  if (ifs) {
    std::string contents;
    ifs.seekg(0, std::ios::end);
    contents.resize(ifs.tellg());
    ifs.seekg(0, std::ios::beg);
    ifs.read(&contents[0], contents.size());
    ifs.close();
    return contents;
  }
  throw(errno);
}

int main(int argc, char *argv[]) {
  // Returns the levenshtein distance between two input files on stdout and the
  // list of repalcement operations
  if (argc != 3) {
    std::cerr << "usage: ./levenshtein bonanza.txt gonzaga.txt" << std::endl;
    return 1;
  }

  std::string a = read(argv[1]);
  std::string b = read(argv[2]);

  std::clock_t start = std::clock();
  std::pair<int, std::vector<cos981::Edit> > de = cos981::levenshtein(a, b);
  std::clock_t end = std::clock();

  int dist = de.first;
  std::vector<cos981::Edit> edits = de.second;
  int elapsed_time = double(end - start) / CLOCKS_PER_SEC;
  std::map<cos981::EditOp, int> s = stats(edits);

  std::string sep;
  for (int i = 0; i < edits.size(); ++i) {
    cos981::EditOp eo = edits[i].second;
    char c = edits[i].first;
    // std::cerr << sep << "(" << c << ", " << to_string(eo) << ")";
    std::cerr << sep << to_string(eo);
    sep = ",";
  }
  std::cerr << std::endl << std::endl;
  std::cerr << "Elapsed Time: " << elapsed_time << std::endl;
  std::cerr << "Statistics: " << to_string(s) << std::endl;
  std::cerr << std::endl << std::endl;

  std::cout << dist << std::endl;

  return 0;
}
